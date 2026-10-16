#!/usr/bin/env python3
"""End-to-end test: real HTTPS agents plus a separate native worker process."""
import hashlib
import json
import pathlib
import ssl
import subprocess
import sys
import tempfile
import time
import urllib.request
import uuid

binary = pathlib.Path(sys.argv[1]).resolve()
with tempfile.TemporaryDirectory(prefix="syncria-worker-smoke-") as work:
    base = pathlib.Path(work)
    processes = []

    def start_agent(name, port):
        root = base / name / "files"
        root.mkdir(parents=True)
        identity = base / name / "identity"
        args = [
            str(binary),
            "-agent",
            "-listen", f"127.0.0.1:{port}",
            "-agent-name", name,
            "-agent-data", str(identity),
            "-agent-root", f"Files={root}",
        ]
        log = open(base / (name + ".log"), "wb")
        process = subprocess.Popen(args, stdout=log, stderr=subprocess.STDOUT)
        processes.append((process, log))
        ctx = ssl._create_unverified_context()
        deadline = time.time() + 20
        while time.time() < deadline:
            if process.poll() is not None:
                raise RuntimeError(f"{name} agent exited unexpectedly")
            if (identity / "identity.json").is_file() and (identity / "agent.crt").is_file():
                try:
                    with urllib.request.urlopen(
                        f"https://127.0.0.1:{port}/v1/status",
                        timeout=1,
                        context=ctx,
                    ) as response:
                        if response.status == 401:
                            break
                except Exception:
                    pass
            time.sleep(0.1)
        else:
            raise RuntimeError(f"agent {name} did not start")
        secret = json.loads((identity / "identity.json").read_text())["secret"]
        pem = (identity / "agent.crt").read_text()
        pin = hashlib.sha256(ssl.PEM_cert_to_DER_cert(pem)).hexdigest()
        record = {
            "id": name,
            "label": name,
            "url": f"https://127.0.0.1:{port}",
            "key": secret,
            "fingerprint": pin,
        }
        return root, record

    def transfer(source, destination, src_file, dst_file, move):
        job = {
            "id": uuid.uuid4().hex,
            "source_id": source["id"],
            "destination_id": destination["id"],
            "source": source["label"],
            "destination": destination["label"],
            "source_root": "Files",
            "destination_root": "Files",
            "file": src_file,
            "target": dst_file,
            "move": move,
            "status": "queued",
        }
        task = {"job": job, "source": source, "destination": destination}
        completed = subprocess.run(
            [str(binary), "-transfer-worker"],
            input=json.dumps(task),
            text=True,
            capture_output=True,
            timeout=45,
        )
        phases = [json.loads(line) for line in completed.stdout.splitlines() if line.strip()]
        assert source["key"] not in completed.stdout, "worker leaked source secret"
        assert destination["key"] not in completed.stdout, "worker leaked destination secret"
        return completed, phases

    try:
        source_root, src = start_agent("source", 19921)
        target_root, dst = start_agent("destination", 19922)
        payload = (b"Syncria independent worker peer transfer\n" * 12345) + b"end"
        (source_root / "first.bin").write_bytes(payload)
        result, phases = transfer(src, dst, "first.bin", "subfolder/first.bin", False)
        assert result.returncode == 0, f"copy failed: {result.stderr} {phases}"
        assert phases[-1]["status"] == "completed", phases
        assert any(p["status"] == "verified" for p in phases), phases
        assert (target_root / "subfolder/first.bin").read_bytes() == payload
        assert (source_root / "first.bin").read_bytes() == payload
        print("PASS: native source -> native destination copy through isolated worker")

        result, phases = transfer(src, dst, "first.bin", "moved.bin", True)
        assert result.returncode == 0, f"move failed: {result.stderr} {phases}"
        assert phases[-1]["status"] == "completed", phases
        assert (target_root / "moved.bin").read_bytes() == payload
        assert not (source_root / "first.bin").exists()
        print("PASS: verified move deletes source only after destination commits")

        (source_root / "second.bin").write_bytes(b"new")
        (target_root / "second.bin").write_bytes(b"original")
        result, phases = transfer(src, dst, "second.bin", "second.bin", True)
        assert result.returncode != 0, "overwriting existing destination succeeded unexpectedly"
        assert phases[-1]["status"] == "failed", phases
        assert (source_root / "second.bin").read_bytes() == b"new"
        assert (target_root / "second.bin").read_bytes() == b"original"
        print("PASS: overwrite rejected; original source preserved")

    finally:
        for process, logfile in processes:
            process.terminate()
        for process, logfile in processes:
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)
            logfile.close()
