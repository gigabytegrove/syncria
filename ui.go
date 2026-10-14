package main

// newUI is a complete replacement for the original single-page dashboard.
const newUI = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Syncria — Your files, in sync</title>
<style>
:root{color-scheme:light;--bg:#f4f8ff;--surface:#ffffff;--surface2:#edf5ff;--line:#d8e4f4;--ink:#092b54;--sub:#526c89;--accent:#0869f8;--accent2:#e6f3ff;--warn:#b56a10;--bad:#b23b42;font:14px/1.55 system-ui,-apple-system,Segoe UI,sans-serif}*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--ink)}button,input,select{font:inherit}button{cursor:pointer}button:disabled{opacity:.45;cursor:not-allowed}a{color:inherit}main{max-width:1500px;margin:auto;display:grid;grid-template-columns:240px minmax(0,1fr);min-height:100vh}.sidebar{position:sticky;top:0;height:100vh;background:#101821;border-right:1px solid var(--line);padding:26px 15px;display:flex;flex-direction:column;gap:25px}.brand{display:flex;align-items:center;gap:12px;padding:0 13px;font-size:22px;font-weight:750;letter-spacing:-1px}.mark{background:var(--accent);color:#091b19;border-radius:12px;width:38px;height:38px;display:grid;place-items:center;font-size:25px}.tiny{font-size:11px;letter-spacing:1.1px;text-transform:uppercase;color:var(--sub);font-weight:750;padding:0 13px}.nav{display:grid;gap:4px}.nav button{display:flex;align-items:center;gap:11px;width:100%;padding:12px 14px;border:1px solid transparent;background:none;text-align:left;border-radius:10px;color:#aabccc}.nav button.active{background:var(--accent2);border-color:#255450;color:#b7f1e7}.nav button:hover{background:#1c2935}.sidebar-bottom{margin-top:auto;padding:0 12px;color:var(--sub);font-size:12px}.content{min-width:0;padding:32px clamp(18px,3.4vw,55px) 70px}.topline{display:flex;justify-content:space-between;align-items:center;gap:18px;margin-bottom:27px}.eyebrow{color:var(--accent);font-size:11px;letter-spacing:1.6px;font-weight:800;text-transform:uppercase}.topline h1{font-size:30px;letter-spacing:-1px;margin:3px 0 0}.lead{color:var(--sub);margin:4px 0 24px}.pill{background:#19302f;color:#88d8c7;padding:6px 11px;border-radius:20px;border:1px solid #2b5852;font-size:12px}.logout{border:1px solid var(--line);background:none;border-radius:9px;color:var(--ink);padding:9px 13px}.view{display:none}.view.active{display:block}.stats{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px;margin-bottom:22px}.stat,.panel{background:var(--surface);border:1px solid var(--line);border-radius:14px}.stat{padding:21px}.stat strong{font-size:31px;font-weight:750;display:block;letter-spacing:-1px}.stat small{color:var(--sub)}.panel{padding:23px;margin-bottom:17px}.panel-head{display:flex;justify-content:space-between;gap:12px;align-items:center;margin-bottom:15px}.panel h2{font-size:17px;letter-spacing:-.4px;margin:0}.panel p{color:var(--sub);margin:5px 0 15px}.split{display:grid;grid-template-columns:1.2fr 1fr;gap:17px}.primary,.secondary,.danger{border-radius:9px;padding:10px 15px;font-weight:700}.primary{background:var(--accent);color:#09201d;border:1px solid var(--accent)}.secondary{background:#1d2a38;border:1px solid #38485a;color:var(--ink)}.danger{background:#3b2428;border:1px solid #704047;color:#ffb8b8}.primary:hover{filter:brightness(1.1)}.row{display:flex;align-items:center;gap:9px;flex-wrap:wrap}.entry{display:flex;align-items:center;justify-content:space-between;gap:16px;border-top:1px solid var(--line);padding:16px 0}.entry:first-of-type{border-top:0}.entry strong{font-size:14px}.entry p{font-size:12px;margin:3px 0}.status{font-size:11px;color:#7ad7ad;border:1px solid #37634c;border-radius:20px;padding:3px 8px}.status.off{color:var(--warn);border-color:#62503b}.empty{padding:30px;text-align:center;border:1px dashed #3c4959;border-radius:11px;color:var(--sub)}.empty strong{display:block;color:var(--ink);margin-bottom:6px}.field{display:block;color:#c2d1df;font-size:12px;font-weight:650;margin:0 0 16px}.field input,.field select{display:block;background:#0e1720;border:1px solid #3b4b5c;border-radius:9px;width:100%;padding:12px;color:var(--ink);margin-top:6px;min-width:0}.field input:focus,.field select:focus{outline:2px solid var(--accent);outline-offset:2px}.muted{color:var(--sub);font-size:12px}.step{display:none}.step.active{display:block}.step-indicator{display:flex;gap:9px;margin-bottom:22px}.step-indicator span{background:#22303c;border:1px solid var(--line);border-radius:9px;padding:8px 12px;color:var(--sub);font-size:12px}.step-indicator span.active{background:var(--accent2);color:#a0e8db;border-color:#29655e}.selected{display:flex;align-items:center;justify-content:space-between;gap:15px;background:#101b26;border:1px solid #354958;border-radius:11px;padding:16px;margin-bottom:18px;min-width:0}.selected strong{display:block;overflow-wrap:anywhere}.picker{position:fixed;inset:0;background:#000b;z-index:30;display:none;align-items:center;justify-content:center;padding:18px}.picker.open{display:flex}.picker-box{background:#182330;border:1px solid #405165;border-radius:16px;width:min(720px,100%);max-height:86vh;display:flex;flex-direction:column;box-shadow:0 25px 90px #000a}.picker-head{padding:22px;border-bottom:1px solid var(--line)}.picker-head h2{margin:0}.picker-path{margin:8px 0 0;font-size:12px;color:var(--sub);overflow-wrap:anywhere}.picker-actions{padding:12px 22px;border-bottom:1px solid var(--line)}.folder-list{padding:9px 14px;min-height:180px;overflow:auto}.folder-row{display:block;width:100%;border:0;background:none;text-align:left;padding:13px 14px;color:var(--ink);border-radius:9px}.folder-row:hover{background:#293747}.picker-footer{border-top:1px solid var(--line);padding:16px 22px;display:flex;justify-content:space-between;gap:10px}.notice{padding:13px 17px;margin-bottom:20px;background:#152f32;border:1px solid #2c6666;border-radius:9px;color:#b3e6e1}.notice.bad{background:#362329;border-color:#774853;color:#ffc3c3}form.inline{display:inline}.settings-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:18px}.hidden,[hidden]{display:none!important}.card-label{color:var(--sub);font-size:12px}code{color:#9fdfd4;overflow-wrap:anywhere}summary{cursor:pointer}.version{color:var(--sub);font-size:12px}@media(max-width:980px){main{grid-template-columns:1fr}.sidebar{position:static;height:auto;padding:13px;gap:12px}.sidebar-bottom,.tiny{display:none}.nav{display:flex;flex-wrap:wrap}.nav button{width:auto;padding:9px 12px}.brand{font-size:19px}.content{padding:22px 16px 60px}}@media(max-width:680px){.stats,.split,.settings-grid{grid-template-columns:1fr}.topline h1{font-size:25px}.panel{padding:17px}.entry{align-items:flex-start;flex-direction:column}.topline{align-items:flex-start}}
/* Syncria official identity: navy wordmark, electric blue and cyan symbol, white canvas. */
body{background:linear-gradient(150deg,#f7fbff 0%,#edf5ff 100%);color:#092b54}
.sidebar{background:#fff;border-right-color:#d6e4f4;box-shadow:3px 0 24px #10396609}
.brand{color:#052b55}.mark{background:linear-gradient(130deg,#0965ff 12%,#07c9eb 58%,#053ecb 100%);color:#fff;box-shadow:0 4px 13px #0965ff30}
.nav button{color:#355576}.nav button.active{color:#0757c7;background:#eaf4ff;border-color:#badcff}.nav button:hover{background:#edf5ff}
.eyebrow{color:#0869f8}.topline h1,.panel h2,.stat strong,.entry strong{color:#092d56}
.pill,.status{background:#ebf4ff;color:#1260c2;border:1px solid #b4d5ff}
.status.off{background:#fff6e8;color:#966113;border-color:#f0d2a0}
.stat,.panel{background:#fff;border:1px solid #d8e4f4;box-shadow:0 8px 28px #16487b09}
.primary{background:linear-gradient(110deg,#0869f8,#098fe9);color:#fff;border:1px solid #0872ef}
.primary:hover{filter:brightness(1.06)}.secondary{background:#f0f6ff;border:1px solid #cfdef2;color:#123d6e}
.logout{color:#123d6e;background:#fff;border-color:#d8e4f4}
.empty{color:#68819d;border-color:#d5e4f3}.empty strong{color:#092d56}
.field{color:#264a70}.field input,.field select{color:#092b54;background:#fff;border-color:#c7d9ec}
.selected{background:#f6faff;border-color:#d7e5f4}.card-label,.muted,.version,.lead{color:#597493}
.step-indicator span{background:#eff5fd;color:#5d7793;border-color:#d5e4f3}.step-indicator span.active{background:#e3f2ff;color:#0965d3;border-color:#a5d2fb}
.picker{background:#092b5470}.picker-box{background:#fff;border-color:#c9dbee;box-shadow:0 25px 90px #092b5440}.folder-row{color:#0c355e}.folder-row:hover{background:#eaf4ff}
.picker-head,.picker-actions,.picker-footer,.entry{border-color:#dce7f5}
.notice{color:#14559a;background:#eaf4ff;border-color:#bddaff}.notice.bad{color:#a82b44;background:#fff0f2;border-color:#e9b6c2}
code{color:#076bd3}
.danger{background:#fff0f2;border:1px solid #f1c2c9;color:#b32238}
.auth-layout{display:block;max-width:600px;margin:0 auto;min-height:100vh}.auth-layout .content{padding:60px 18px}.auth-brand{display:block;width:220px;max-width:100%;margin:0 0 42px}.auth-layout .topline{margin-bottom:25px}.auth-layout .panel{max-width:100%!important}.stats .stat strong{font-variant-numeric:tabular-nums}.stats-compact{grid-template-columns:repeat(4,minmax(0,1fr))}.stats-compact .stat{padding:12px}.stats-compact .stat strong{font-size:20px}.wizard-status{padding:12px 14px;border-radius:9px;background:#f2f7fd;border:1px solid #d5e4f3;margin:12px 0;color:#22486a}.wizard-status.good{background:#e8f7f2;border-color:#b9e6d6;color:#12634d}.wizard-status.bad{background:#fff0f1;border-color:#f0bdc4;color:#9e283f}.step-indicator{flex-wrap:wrap}.review-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}.review-grid>div{padding:13px;background:#f5faff;border:1px solid #dae7f4;border-radius:10px}.review-grid strong{display:block;overflow-wrap:anywhere}.mini-stat{font-variant-numeric:tabular-nums}@media(max-width:680px){.stats-compact,.review-grid{grid-template-columns:1fr 1fr}}
.update-progress-panel{margin:20px 0 4px;padding:18px;border-radius:12px;border:1px solid #c4daf5;background:#f6faff}
.update-progress-panel progress{width:100%;height:15px;accent-color:#0869f8;display:block;margin:13px 0 9px}
.update-progress-panel .update-progress-meta{display:flex;justify-content:space-between;gap:10px;align-items:baseline;flex-wrap:wrap}
.update-progress-panel .update-steps{list-style:none;display:flex;gap:8px;flex-wrap:wrap;padding:0;margin:15px 0 0}
.update-progress-panel .update-steps li{border:1px solid #d6e3f5;padding:5px 10px;border-radius:7px;color:#597493;font-size:12px}
.update-progress-panel .update-steps li.active{background:#e3f2ff;color:#0869f8;border-color:#91bbef;font-weight:700}
.update-progress-panel .update-steps li.done{color:#146c52;background:#e7f6ed;border-color:#bcdfc9}
.update-progress-panel .update-steps li.failed{color:#a02835;background:#fff2f4;border-color:#e6aeb8}
` + v051Theme + `</style></head><body>
<main class="{{if and .Configured .Logged}}app-layout{{else}}auth-layout{{end}}">
{{if and .Configured .Logged}}
<aside class="sidebar" aria-label="Syncria navigation">
  <div class="sidebar-brand"><div class="sidebar-logo"><img src="/assets/syncria-logo.svg" alt="Syncria" width="196" height="70"></div><div class="sidebar-edition"><span class="pulse-dot"></span> SYNC WORKSPACE <span class="sidebar-edition-line"></span></div></div>
  <div class="sidebar-navigation">
    <div class="nav-group"><p class="nav-section-label">WORKSPACE</p><nav class="nav" aria-label="Workspace">
      <button type="button" data-view="overview" class="active" aria-label="Overview"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="7" height="7" rx="1.6"/><rect x="14" y="3" width="7" height="7" rx="1.6"/><rect x="3" y="14" width="7" height="7" rx="1.6"/><rect x="14" y="14" width="7" height="7" rx="1.6"/></svg></span><span class="nav-copy"><strong>Overview</strong><small>Your files at a glance</small></span></button>
      <button type="button" data-view="sync" class="" aria-label="My Syncs"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 7h15m0 0-4-4m4 4-4 4M20 17H5m0 0 4-4m-4 4 4 4"/></svg></span><span class="nav-copy"><strong>My Syncs</strong><small>Manage file flows</small></span></button>
      <button type="button" data-view="activity" class="" aria-label="Activity"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 17h4l3-10 4 13 3-7h4"/></svg></span><span class="nav-copy"><strong>Activity</strong><small>Transfers and results</small></span></button>
    </nav></div>
    <div class="nav-group"><p class="nav-section-label">CONNECTIONS</p><nav class="nav" aria-label="Connections">
      <button type="button" data-view="storage" class="" aria-label="My NAS Boxes"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="8" rx="2"/><rect x="3" y="13" width="18" height="8" rx="2"/><path d="M7 7h.01M7 17h.01M11 7h6M11 17h6"/></svg></span><span class="nav-copy"><strong>My NAS Boxes</strong><small>Network storage</small></span></button>
      <button type="button" data-view="accounts" class="" aria-label="Cloud Accounts"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 16.5a4 4 0 0 1 1-7.7 7.5 7.5 0 0 1 14.4 1.2A4.5 4.5 0 0 1 18 19H6.5"/><path d="M8 16h8m-4-4v8"/></svg></span><span class="nav-copy"><strong>Cloud Accounts</strong><small>Google & Dropbox</small></span></button>
    </nav></div>
    <div class="nav-group"><p class="nav-section-label">SYSTEM</p><nav class="nav" aria-label="System">
      <button type="button" data-view="updates" class="" aria-label="App Updates"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4.8 9a7.5 7.5 0 0 1 13.4-3L21 9m0-6v6h-6M19.2 15A7.5 7.5 0 0 1 5.8 18L3 15m0 6v-6h6"/></svg></span><span class="nav-copy"><strong>App Updates</strong><small>Versions and installs</small></span></button>
      <button type="button" data-view="settings" class="" aria-label="Settings"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 3 14 5l3-1 1 3 3 2-2 3 2 3-3 2-1 3-3-1-2 2-2-2-3 1-1-3-3-2 2-3-2-3 3-2 1-3 3 1z"/><circle cx="12" cy="12" r="3"/></svg></span><span class="nav-copy"><strong>Settings</strong><small>Advanced configuration</small></span></button>
    </nav></div>
  </div>
  <div class="sidebar-bottom"><div class="sidebar-presence"><span class="sidebar-presence-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 7h15m0 0-4-4m4 4-4 4M20 17H5m0 0 4-4m-4 4 4 4"/></svg></span><span><strong>Syncria</strong><small>Your files stay yours.</small></span></div><div class="sidebar-version">VERSION <strong><span id="sidebarVersion">{{.Version}}</span></strong></div></div>
</aside>
{{end}}
<div class="content">{{if not .Logged}}<img class="auth-brand" src="/assets/syncria-logo.svg" alt="Syncria">{{end}}
<header class="topline"><div class="page-heading"><div class="eyebrow"><span class="breadcrumb-home">SYNC WORKSPACE</span><span class="breadcrumb-divider">/</span><span id="pageBreadcrumb">{{if not .Configured}}GETTING STARTED{{else if not .Logged}}WELCOME BACK{{else}}OVERVIEW{{end}}</span></div><h1 id="pageTitle">{{if not .Configured}}Welcome to Syncria{{else if not .Logged}}Sign in{{else}}Your overview{{end}}</h1></div><div class="header-actions">{{if .Logged}}<span class="header-version">Syncria {{.Version}}</span><form method="post" action="/action"><input type="hidden" name="action" value="logout"><button class="logout" type="submit">Sign out ↗</button></form>{{end}}</div></header>
{{if .Message}}<div class="notice">{{.Message}}</div>{{end}}{{if and .Configured .Logged}}<div id="updateLiveBanner" class="update-live-banner hidden" role="status"><div><strong>Update running in Syncria</strong><div><small id="updateLiveBannerStatus">Checking the updater status…</small></div></div><button type="button" class="secondary" onclick="navigate('updates')">View exact progress →</button></div>{{end}}{{if and .Configured .Logged}}<div id="sessionExpiredNotice" class="notice bad hidden" role="alert">Your Syncria session has expired. <a href="/">Sign in again</a> to continue.</div>{{end}}
{{if not .Configured}}<section class="panel" style="max-width:550px"><h2>Welcome to Syncria</h2><p>Create the administrator password to protect your dashboard.</p><form method="post" action="/action"><input type="hidden" name="action" value="setup"><label class="field">Administrator password<input type="password" name="password" minlength="12" required autocomplete="new-password"></label><button class="primary">Secure dashboard</button></form></section>
{{else if not .Logged}}<section class="panel" style="max-width:550px"><h2>Sign in</h2><p>Access your synchronization workspace.</p><form method="post" action="/action"><input type="hidden" name="action" value="login"><label class="field">Administrator password<input type="password" name="password" required autocomplete="current-password"></label><button class="primary">Sign in</button></form></section>
{{else}}
<section class="view active" id="view-overview">
<div class="command-hero">
  <div class="hero-text"><div class="hero-overline"><span class="hero-overline-dot"></span> YOUR FILES. CONNECTED.</div><h2>Everything in sync.<br><em>Wherever it lives.</em></h2><p>Google Drive, Dropbox, NAS boxes. Syncria keeps the folders you choose together without keeping your files on this server.</p><div class="hero-actions"><button class="hero-cta" type="button" onclick="startWizard()">Create a sync <span aria-hidden="true">↗</span></button><button class="hero-text-link" type="button" onclick="navigate('sync')">Explore my syncs <span aria-hidden="true">→</span></button></div></div>
  <div class="hero-illustration" aria-hidden="true">
    <svg class="hero-flow-art" viewBox="0 0 420 255" fill="none" xmlns="http://www.w3.org/2000/svg">
     <path class="hero-path" d="M75 128C140 128 140 70 210 70C280 70 279 128 345 128" stroke="currentColor" stroke-width="2" stroke-dasharray="5 9"/>
     <path class="hero-path hero-path-low" d="M75 128C140 128 140 187 210 187C280 187 279 128 345 128" stroke="currentColor" stroke-width="2" stroke-dasharray="5 9"/>
     <circle cx="75" cy="128" r="37" fill="#142f61" stroke="#55a2fb" stroke-width="1.5"/>
     <path d="M55 133a17 17 0 0 1 2-23 18 18 0 0 1 34 6 13 13 0 0 1-2 26H60a12 12 0 0 1-5-9Z" stroke="#9dd8ff" stroke-width="3" stroke-linejoin="round"/>
     <circle cx="210" cy="70" r="31" fill="#18d3d2"/><path d="m195 72 10 10 20-23" stroke="#0b2549" stroke-width="5" stroke-linecap="round" stroke-linejoin="round"/>
     <circle cx="210" cy="187" r="31" fill="#123d78" stroke="#6fa9ec" stroke-width="1.5"/>
     <path d="M196 179h28v21h-28zM201 184h18M201 190h18M201 196h9" stroke="#a8dcff" stroke-width="2" stroke-linejoin="round"/>
     <circle cx="345" cy="128" r="37" fill="#142f61" stroke="#55a2fb" stroke-width="1.5"/>
     <rect x="324" y="113" width="42" height="12" rx="3" stroke="#b7e4ff" stroke-width="2.5"/><rect x="324" y="131" width="42" height="12" rx="3" stroke="#b7e4ff" stroke-width="2.5"/><circle cx="332" cy="119" r="2" fill="#18d3d2"/><circle cx="332" cy="137" r="2" fill="#18d3d2"/>
     <text x="75" y="188" text-anchor="middle" fill="#cfe7ff" font-size="10" letter-spacing="1.5">CLOUD</text>
     <text x="345" y="188" text-anchor="middle" fill="#cfe7ff" font-size="10" letter-spacing="1.5">NAS</text>
     <text x="210" y="24" text-anchor="middle" fill="#cfe7ff" font-size="10" letter-spacing="1.5">IN SYNC</text>
    </svg><div class="hero-flow-foot">CLOUD <span>↔</span> SYNCRIA <span>↔</span> YOUR STORAGE</div>
  </div>
</div>
<div class="overview-metrics" aria-label="Your current Syncria numbers">
 <div class="overview-metric"><span class="metric-symbol">↗</span><div><small>YOUR SYNCS</small><strong>{{len .Jobs}}</strong><span>configured connections</span></div></div>
 <div class="overview-metric"><span class="metric-symbol">▦</span><div><small>FILES TRACKED</small><strong id="metricFiles">{{.Metrics.TrackedFiles}}</strong><span>across your syncs</span></div></div>
 <div class="overview-metric"><span class="metric-symbol">⇅</span><div><small>DATA MOVED</small><strong id="metricBandwidth">—</strong><span>total transferred</span></div></div>
 <div class="overview-metric"><span class="metric-symbol">✓</span><div><small>SYNC RUNS</small><strong id="metricRuns">{{.Metrics.Runs}}</strong><span><span id="metricFailures">{{.Metrics.Failures}}</span> with errors</span></div></div>
</div>
{{if not .Jobs}}
<div class="welcome-track"><div class="welcome-track-intro"><span class="mini-overline">YOUR FIRST SYNC</span><h2>Make your first connection.</h2><p>New here? Start with these three steps. You only need to set this up once.</p></div>
<div class="welcome-track-steps">
 <div><span class="track-index">01</span><strong>Connect your cloud</strong><p>Google Drive or Dropbox. Skip this if you're using two NAS boxes.</p><button type="button" onclick="startAccountWizard()">Cloud accounts →</button></div>
 <div><span class="track-index">02</span><strong>Add a NAS box</strong><p>Tell Syncria where your storage box lives.</p><button type="button" onclick="toggleDirectNAS(true)">Storage →</button></div>
 <div><span class="track-index">03</span><strong>Choose your folders</strong><p>Pick which folders should stay matched. New syncs start off.</p><button type="button" onclick="startWizard()">Set up sync →</button></div>
</div></div>
{{else}}
<details class="onboarding-disclosure"><summary>Need help setting up another sync? Open the step-by-step guide</summary><div class="quick-start"><div class="quick-start-heading"><h2>New here? Start with these three steps.</h2></div><div class="quick-start-grid">
<div class="quick-step"><span class="step-number">1</span><h3>Connect your cloud files</h3><p>Google Drive or Dropbox, if you use them.</p><button class="secondary" type="button" onclick="startAccountWizard()">Connect account</button></div>
<div class="quick-step"><span class="step-number">2</span><h3>Add your NAS box</h3><p>Give Syncria your NAS address and sign-in details.</p><button class="secondary" type="button" onclick="toggleDirectNAS(true)">Add NAS</button></div>
<div class="quick-step"><span class="step-number">3</span><h3>Choose where files go</h3><p>Choose the folders to match.</p><button class="primary" type="button" onclick="startWizard()">Set up a sync</button></div>
</div></div></details>
{{end}}
<div class="overview-columns">
<section class="overview-list-panel">
 <div class="section-heading"><div><span class="mini-overline">01 / YOUR FLOWS</span><h2>My Syncs</h2><p>Each sync is a connection between places you've chosen.</p></div><button class="text-action" type="button" onclick="navigate('sync')">View all →</button></div>
 <div class="flow-list">
 {{range .Jobs}}<div class="flow-row"><span class="flow-row-icon">{{if eq .Provider "bridge"}}☁{{else}}⇄{{end}}</span><div class="flow-row-content"><strong>{{.Name}}</strong><small>{{if eq .Provider "bridge"}}{{if eq .CloudProvider "dropbox"}}Dropbox{{else}}Google Drive{{end}} → NAS{{else if eq .Provider "nas_bridge"}}NAS ↔ NAS{{else}}Existing folder sync{{end}}</small></div>{{if .Enabled}}<span class="flow-state is-on"><span></span> On</span>{{else}}<span class="flow-state is-off"><span></span> Paused</span>{{end}}<button class="flow-row-open" type="button" onclick="navigate('sync')" aria-label="Open {{.Name}}">↗</button></div>{{else}}<div class="overview-empty"><span>⇄</span><strong>No syncs yet</strong><p>Make your first one to start keeping folders together.</p><button class="secondary" type="button" onclick="startWizard()">Create a sync</button></div>{{end}}
 </div>
</section>
<section class="overview-list-panel">
 <div class="section-heading"><div><span class="mini-overline">02 / YOUR PLACES</span><h2>Connected storage</h2><p>Cloud accounts and NAS boxes saved in Syncria.</p></div><button class="text-action" type="button" onclick="navigate('storage')">Manage →</button></div>
 <div class="place-summary"><div class="place-summary-icon cloud-symbol">☁</div><div><strong>Cloud accounts</strong><small>Google Drive: {{len .Accounts}} · Dropbox: {{len .DropboxAccounts}}</small></div><button class="flow-row-open" type="button" onclick="navigate('accounts')" aria-label="Open Cloud Accounts">↗</button></div>
 {{range .Shares}}<div class="place-summary"><div class="place-summary-icon">▤</div><div><strong>{{.Label}}</strong><small>{{.Protocol}} · {{.Server}}</small></div>{{if not .Local}}<span class="place-indicator" data-share-mounted="{{.ID}}">NAS connection saved</span>{{else if .Mounted}}<span class="place-indicator" data-share-mounted="{{.ID}}">Mounted</span>{{else}}<span class="place-indicator" data-share-mounted="{{.ID}}">Not mounted</span>{{end}}</div>{{else}}<div class="overview-empty overview-empty-compact"><strong>No NAS boxes connected</strong><p>Add a NAS from Storage to use it in a sync.</p></div>{{end}}
 </section>
</div>
<details class="glossary-fold"><summary>New to syncing? A few words explained.</summary><div class="glossary-layout"><span><strong>NAS</strong> — A NAS is a storage box on your network.</span><span><strong>Cloud</strong> — Online files, like Google Drive.</span><span><strong>Sync</strong> — Keeping folders matched.</span><span><strong>Undo</strong> — Bringing back a recently deleted file.</span></div></details>
</section>
<section class="view" id="view-sync">
<div class="workspace-intro"><div><span class="mini-overline">FLOWS / YOUR CONNECTIONS</span><h2>Every sync, one place.</h2><p>Manage where files travel and when they move. Nothing happens automatically until you turn a sync on.</p></div><button class="primary" type="button" onclick="startWizard()">+ Create sync</button></div>
<section class="sync-library" aria-label="Your file syncs">
{{range .Jobs}}{{$job := .}}
<article class="relationship-card">
 <div class="flow-card-main">
   <div class="flow-card-emblem">{{if eq .Provider "bridge"}}☁{{else}}⇄{{end}}</div>
   <div class="flow-card-identity"><span class="mini-overline">{{if eq .Provider "bridge"}}CLOUD / NAS{{else if eq .Provider "nas_bridge"}}NAS / NAS{{else}}FOLDER / FOLDER{{end}}</span><h3>{{.Name}}</h3><p>{{if eq .Provider "bridge"}}{{if eq .CloudProvider "dropbox"}}Dropbox{{else}}Google Drive{{end}} and {{len .NASDestinations}} NAS destination(s){{else if eq .Provider "nas_bridge"}}{{if eq .Direction "one_way"}}Copy one way{{else}}Keep both ways matched{{end}} · {{len .NASDestinations}} NAS destination(s){{else}}Existing filesystem sync{{end}}</p></div>
   <div class="flow-card-health">{{if .Enabled}}<span class="flow-state is-on" data-job-enabled="{{.ID}}"><span></span> On</span>{{else}}<span class="flow-state is-off" data-job-enabled="{{.ID}}"><span></span> Paused</span>{{end}}</div>
 </div>
 <div class="flow-card-facts">
    <div><small>WHEN IT CHECKS</small><strong>{{if eq .ScheduleMode "changes"}}On NAS changes{{else}}Every {{.Interval}} minutes{{end}}</strong></div>
    <div><small>LAST RUN</small><strong data-job-run="{{.ID}}">{{if .LastRun.IsZero}}Not yet{{else}}{{.LastRun.Format "Jan 2, 2006 · 15:04"}}{{end}}</strong></div>
    <div><small>YOUR FILES</small><strong>{{len .NASDestinations}} NAS folder(s)</strong></div>
 </div>
 <p class="flow-card-description" data-job-summary="{{.ID}}">{{if .LastSummary}}{{.LastSummary}}{{else}}This sync hasn't run yet.{{end}}</p>
 <p class="flow-card-stats" data-job-statline="{{.ID}}">Sent: {{.TotalUploads}} files · Received: {{.TotalDownloads}} files · Successful runs: {{.SuccessfulRuns}} · Errors: {{.FailedRuns}}</p>
 <p class="flow-card-error" data-job-error="{{.ID}}" role="alert" {{if not .LastError}}hidden{{end}}>{{.LastError}}</p>
 <div class="sync-live-progress hidden" data-job-progress="{{.ID}}">
  <div class="progress-topline"><strong data-job-phase="{{.ID}}">Syncing…</strong><small>Live transfer progress</small></div>
  <progress data-job-bar="{{.ID}}" max="100"></progress><p data-job-progress-detail="{{.ID}}"></p>
 </div>
 {{if .Conflicts}}<section class="relationship-alert" aria-label="Files that need a decision"><div class="alert-title"><span>!</span><div><h4>Two versions need your choice</h4><p>Syncria kept both copies safe until you choose.</p></div></div>
 {{range .Conflicts}}<div class="decision-row"><div><strong>{{.Path}}</strong><small>First NAS: {{.PrimaryModified.Format "Jan 2, 2006 · 15:04"}} · Other NAS: {{.SecondaryModified.Format "Jan 2, 2006 · 15:04"}}</small></div>
 <form method="post" action="/api/conflicts/resolve" class="inline"><input type="hidden" name="id" value="{{$job.ID}}"><input type="hidden" name="pair" value="{{.Pair}}"><input type="hidden" name="path" value="{{.Path}}">
 <label class="field"><span class="sr-only">Which copy to keep</span><select name="choice"><option value="primary">Use first NAS copy</option><option value="secondary">Use other NAS copy</option><option value="keep_both">Keep both copies</option></select></label><button type="submit" class="secondary">Save choice</button></form></div>{{end}}
 </section>{{end}}
 {{if .PendingDeletions}}<section class="relationship-alert recovery-alert" aria-label="Files you can restore"><div class="alert-title"><span>↶</span><div><h4>Recently deleted · 10-minute Undo</h4><p>These files can be brought back before their recovery time runs out.</p></div></div>
 {{range .PendingDeletions}}<div class="decision-row"><div><strong>{{.Path}}</strong><small>Undo before {{.Expires.Format "Jan 2, 2006 · 15:04:05 MST"}}</small></div>
 <form method="post" action="/api/recovery/undo" class="inline"><input type="hidden" name="operation" value="undo"><input type="hidden" name="id" value="{{$job.ID}}"><input type="hidden" name="pair" value="{{.Pair}}"><input type="hidden" name="path" value="{{.Path}}"><button type="submit" class="secondary">Bring back file</button></form></div>{{end}}
 </section>{{end}}
 <div class="flow-card-actions">
 <form method="post" action="/action" class="inline"><input type="hidden" name="action" value="run_job"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="primary">Sync now <span aria-hidden="true">→</span></button></form>
 <form method="post" action="/action" class="inline"><input type="hidden" name="action" value="toggle_job"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="secondary">{{if .Enabled}}Pause automatic sync{{else}}Turn on automatic sync{{end}}</button></form>
 <form method="post" action="/action" class="inline flow-danger" onsubmit="return confirm('Remove this sync? Files stay on your cloud and NAS.')"><input type="hidden" name="action" value="delete_job"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="danger">Remove sync</button></form>
 </div>
</article>
{{else}}
<div class="flow-library-empty"><span class="empty-symbol">⇄</span><span class="mini-overline">NO FLOWS YET</span><h3>Your first sync starts here.</h3><p>Connect two folders, then decide whether and when Syncria should keep them matched.</p><button type="button" class="primary" onclick="startWizard()">Create my first sync →</button></div>
{{end}}
</section>
<details class="glossary-fold"><summary>Questions about one-way, both ways, or Undo?</summary><div class="glossary-layout"><span><strong>One way</strong> copies changes from the first NAS to the others.</span><span><strong>Both ways</strong> lets changes move between NAS boxes.</span><span><strong>Undo</strong> can bring back a recent deletion.</span></div></details>
<section id="wizard" class="sync-builder hidden"><div class="builder-heading"><div><span class="mini-overline">SYNC BUILDER / NEW FLOW</span><h2>Connect your files.</h2><p>Four short steps. No files move until you're ready.</p></div><button type="button" class="secondary" onclick="closeSyncWizard()">Close ✕</button></div>
<div class="step-indicator"><span id="stepMarker1" class="active">1 · Files start here</span><span id="stepMarker2">2 · Copy them here</span><span id="stepMarker3">3 · Pick the rules</span><span id="stepMarker4">4 · Double-check</span></div>
<form id="syncWizard" method="post" action="/action" onsubmit="return submitSyncWizard(event)" novalidate><input type="hidden" name="action" value="add_bridge">
<div id="wizard1" class="step active"><h2>Step 1: Where are your files now?</h2><p class="muted">Choose the place that already has your files. We won't move or delete anything while you set this up.</p>
<label class="field">Give this sync a name <span class="field-help">Pick any name you will remember, like “Family Photos”.</span><input name="name" id="syncName" maxlength="128" placeholder="Family Photos" required></label>
<label class="field">What do you want to keep in sync? <span class="field-help">A NAS is a file storage box connected to your home or work network.</span><select id="relationshipType" name="relationship_type" onchange="relationshipTypeChanged()"><option value="cloud">My cloud files and a NAS box</option><option value="nas">Two NAS boxes</option></select></label>
<div id="cloudSourceFields">
<label class="field">Where are your cloud files? <span class="field-help">Choose the service that already has your files.</span><select id="sourceProvider" name="cloud_provider" onchange="sourceProviderChanged()"><option value="google_drive">Google Drive</option><option value="dropbox">Dropbox</option></select></label>
<div id="googleSourceFields"><label class="field">Which Google Drive account? <span class="field-help">Pick the email address that owns the files.</span><select id="syncAccount" name="account" onchange="resetDrive()"><option value="">Select a Google account</option>{{range .Accounts}}<option value="{{.ID}}">{{.Email}}</option>{{end}}</select></label></div>
<div class="hidden" id="dropboxSourceFields"><label class="field">Which Dropbox account? <span class="field-help">Pick the email address that owns the files.</span><select id="syncDropboxAccount" name="dropbox_account" onchange="resetDrive()"><option value="">Select a Dropbox account</option>{{range .DropboxAccounts}}<option value="{{.ID}}">{{.Email}}</option>{{end}}</select></label></div>
<div class="selected"><div><span class="card-label">Which cloud folder?</span><strong id="selectedDriveLabel">Entire My Drive</strong></div><button class="secondary" type="button" onclick="openPicker($('sourceProvider').value==='dropbox'?'dropbox':'drive')">Choose folder</button></div>
<input type="hidden" name="remote" id="remoteValue" value="root"></div>
<div class="hidden" id="nasSourceFields"><label class="field">Which NAS box has your files? <span class="field-help">We call this your first NAS. You can choose another NAS for the copies in Step 2.</span><select id="nasSourceId" name="nas_source_id" onchange="nasSourceChanged()"><option value="">Choose the source NAS</option>{{range .Shares}}{{if not .Local}}{{if eq .Protocol "smb"}}<option value="{{.ID}}">{{.Label}} — {{.Server}} / {{.Export}}</option>{{end}}{{end}}{{end}}</select></label>
<div class="selected"><div><span class="card-label">Folder on your first NAS</span><strong id="selectedNASSourceLabel">Share root</strong></div><button class="secondary" type="button" onclick="openPicker('nas-source')">Browse folder</button></div>
<input type="hidden" id="nasSourceFolderValue" name="nas_source_folder" value="."></div>
<p class="wizard-status" id="syncWizardStatus" role="status">Pick where your files are. Syncria will check that it can reach them.</p>
<div class="row"><button id="verifySourceButton" class="primary" type="button" onclick="wizardNext()">Next: where should copies go? →</button><button type="button" class="secondary" onclick="navigate($('relationshipType').value==='cloud'?'accounts':'storage')">Need to add an account or NAS?</button></div></div>
<div id="wizard2" class="step"><h2>Step 2: Where should copies go?</h2><p class="muted">Choose at least one NAS box and its folder. You can choose more than one. The files stay on the NAS boxes, not inside Syncria.</p>
<div id="nasTargets">
{{range .Shares}}{{if not .Local}}{{if eq .Protocol "smb"}}
<div class="entry nas-target"><div><label><input type="checkbox" name="nas_ids" value="{{.ID}}" onchange="nasDestChanged()"> <strong>{{.Label}}</strong></label><p>{{.Server}} / {{.Export}}</p><small id="nas-label-{{.ID}}">Share root</small></div>
<div><button type="button" class="secondary" onclick="browseNASDestination('{{.ID}}')">Choose folder</button><input type="hidden" id="nas-folder-{{.ID}}" name="nas_folder_{{.ID}}" value="."></div></div>
{{end}}{{end}}{{end}}
</div><p class="wizard-status" id="nasDestStatus">Choose one or more NAS boxes for your files.</p>
<div class="row"><button type="button" class="secondary" onclick="showSyncStep(1)">← Back</button><button id="verifyTargetButton" type="button" class="primary" onclick="wizardReview()">Next: choose the rules →</button><button type="button" class="secondary" onclick="toggleDirectNAS(true)">Add another NAS box</button></div></div>
<div id="wizard3" class="step"><h2>Step 3: How should this sync work?</h2><p class="muted">The usual choices are already picked for you. You can change them whenever you set up a new sync.</p>
<label class="field">How often should Syncria check your files? <span class="field-help">Checking looks for files that are new, changed, or missing.</span><select id="syncInterval" name="interval"><option value="5">Every 5 minutes</option><option value="10">Every 10 minutes</option><option value="15">Every 15 minutes</option><option value="30">Every 30 minutes</option><option value="60">Every hour</option><option value="360">Every 6 hours</option><option value="1440">Daily</option></select></label>
<details class="simple-details"><summary>More choices about when to check</summary><label class="field">When should Syncria check? <span class="field-help">“When a file changes” asks your NAS to tell Syncria when it notices changes. Cloud changes are also checked every 5 minutes.</span><select id="scheduleMode" name="schedule_mode" onchange="scheduleModeChanged()"><option value="interval">Every few minutes (recommended)</option><option value="changes">When a NAS file changes, plus regular checks</option></select></label></details>
<div id="nasDirectionFields" class="hidden"><label class="field">Which way should files go? <span class="field-help">One-way copies from the first NAS to the others. Both ways lets changes travel between them.</span><select id="nasDirection" name="direction" onchange="nasDirectionChanged()"><option value="one_way">One way: first NAS → other NAS</option><option value="bidirectional">Both ways: keep both NAS boxes matched</option></select></label>
<div id="nasConflictFields" class="hidden"><label class="field">What if the same file was changed in both places? <span class="field-help">This only matters when both NAS boxes have different copies of one file.</span><select name="conflict_policy" id="conflictPolicy"><option value="manual">Ask me which copy to keep (safest)</option><option value="newer">Use the copy changed most recently</option><option value="primary">Always choose the first NAS copy</option><option value="secondary">Always choose the other NAS copy</option><option value="keep_both">Save both copies with different names</option></select></label></div></div>
<details id="crossNASField" class="simple-details"><summary>Extra choice: let several NAS boxes share changes through the cloud</summary><label><input type="checkbox" name="cross_nas_propagation" value="1"> Let changes on one NAS reach my other NAS boxes through the cloud</label><p class="muted">Off by default. If you set up a separate NAS-to-NAS sync, Syncria uses that instead to share files between your boxes.</p></details>
<div class="panel" style="margin-top:16px"><h3>What happens when I delete a file?</h3>
<label><input id="propagateDeletions" type="checkbox" name="propagate_deletions" value="1" checked> Delete the matching copies too</label>
<p class="muted">On by default. For example, deleting a file from one NAS also deletes its matching copy on the other NAS. Turn this off if you want the other copy to stay.</p>
<label><input id="recoveryEnabled" type="checkbox" name="recovery_enabled" value="1" checked> Give me 10 minutes to undo an accidental deletion</label>
<p class="muted">On by default. After Syncria notices a deleted file, you get 10 minutes to bring it back. The saved file stays on your NAS or cloud service.</p></div>
<div class="row"><button type="button" class="secondary" onclick="showSyncStep(2)">← Back</button><button type="button" class="primary" onclick="wizardFinalize()">Next: check my choices →</button></div></div>
<div id="wizard4" class="step"><h2>Step 4: Does this look right?</h2><p class="muted">Read these choices before saving. You can go back and change any of them.</p>
<div class="review-grid"><div><span class="card-label">Sync name</span><strong id="reviewName"></strong></div>
<div><span class="card-label">Files come from</span><strong id="reviewProvider"></strong><small id="reviewSource"></small></div>
<div><span class="card-label">Files go to</span><strong id="reviewTarget"></strong></div>
<div><span class="card-label">When Syncria checks</span><strong id="reviewInterval"></strong></div>
<div><span class="card-label">How files are matched</span><strong id="reviewDirection"></strong></div>
<div><span class="card-label">Deleting and bringing back files</span><strong id="reviewDeletion"></strong></div></div>
<p class="wizard-status good" id="reviewStatus">Your new sync starts turned off. Files will not be copied until you click Turn on sync or Sync now.</p>
<div class="row"><button type="button" class="secondary" onclick="showSyncStep(3)">← Back</button><button type="submit" class="primary" id="createSyncButton">Save my sync (keep it off)</button></div></div></form></section></section>
<section class="view" id="view-storage">
<div class="workspace-intro"><div><span class="mini-overline">CONNECTIONS / NETWORK STORAGE</span><h2>Give your files another home.</h2><p>Add NAS boxes here, then use their folders in any sync. Syncria doesn't move your files just by connecting a NAS.</p></div><button class="primary" type="button" onclick="toggleDirectNAS()">+ Add NAS</button></div>
<section id="directNASPanel" class="connection-drawer hidden" aria-label="Add a NAS connection"><div class="drawer-heading"><div><span class="mini-overline">NEW CONNECTION</span><h3>Connect a NAS box</h3><p>You'll need the address, shared folder, and sign-in details from your NAS.</p></div><button type="button" class="secondary" onclick="toggleDirectNAS(false)">Close ✕</button></div>
<form method="post" action="/api/nas/action"><input type="hidden" name="operation" value="connect">
<div class="settings-grid">
<label class="field">Give it a name<span class="field-help">Something you'll recognize, like “Office NAS”.</span><input name="label" id="shareLabel" placeholder="Family NAS" maxlength="128" required></label>
<label class="field">How does it share files?<span class="field-help">The default works for most NAS boxes.</span><select name="protocol" id="shareProtocol"><option value="smb">Windows-style shared folders (SMB) — recommended</option><option value="nfs" disabled>Other network sharing (NFS) — not available yet</option></select></label>
<label class="field">Network address<span class="field-help">Usually looks like 192.168.1.20.</span><input name="server" id="shareServer" placeholder="192.168.30.10" required></label>
<label class="field">Shared folder<span class="field-help">The folder name your NAS makes available.</span><input name="export" id="sharePath" placeholder="Documents" required></label>
<label class="field">NAS username<input name="username" required autocomplete="username"></label>
<label class="field">NAS password<input type="password" name="password" required autocomplete="new-password"></label>
</div><div class="drawer-footer"><p>Syncria checks the connection before saving it.</p><button class="primary" type="submit">Test & save connection <span aria-hidden="true">→</span></button></div></form></section>
<section class="connection-list"><div class="section-heading"><div><span class="mini-overline">YOUR STORAGE</span><h2>My NAS Boxes</h2><p>Each saved connection can be used by one or more syncs.</p></div><span class="section-counter">{{len .Shares}} saved</span></div>
{{range .Shares}}<div class="connection-row"><div class="connection-emblem"><svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.7"><rect x="3" y="3" width="18" height="8" rx="2"/><rect x="3" y="13" width="18" height="8" rx="2"/><path d="M7 7h.01M7 17h.01"/></svg></div><div class="connection-copy"><strong>{{.Label}}</strong><p>{{.Server}} / {{.Export}}</p><small>{{if .Local}}Older mounted connection — kept as configured{{else}}Direct SMB connection{{end}}</small></div><div class="connection-tools">
{{if not .Local}}<span class="connection-state" data-share-mounted="{{.ID}}">Saved connection</span>
<form class="inline" action="/api/nas/action" method="post"><input type="hidden" name="operation" value="verify"><input type="hidden" name="id" value="{{.ID}}"><button class="secondary" type="submit">Test</button></form>
<form class="inline" action="/api/nas/action" method="post" onsubmit="return confirm('Remove this NAS connection? Files on the NAS will not be touched.')"><input type="hidden" name="operation" value="disconnect"><input type="hidden" name="id" value="{{.ID}}"><button class="danger" type="submit">Remove</button></form>{{else}}<span class="connection-state" data-share-mounted="{{.ID}}">{{if .Mounted}}Mounted{{else}}Not mounted{{end}}</span>{{end}}</div></div>
{{else}}<div class="flow-library-empty"><span class="empty-symbol">▤</span><span class="mini-overline">NOT CONNECTED YET</span><h3>Your NAS boxes will appear here.</h3><p>Add one to get started. Files won't be copied until you set up a sync.</p><button type="button" class="primary" onclick="toggleDirectNAS(true)">Add a NAS →</button></div>{{end}}</section>
</section>
<section class="view" id="view-accounts">
<div class="workspace-intro"><div><span class="mini-overline">CONNECTIONS / CLOUD</span><h2>One home for all your clouds.</h2><p>Connect the accounts you want to use. Each account stays separate, and nothing syncs until you choose a folder.</p></div><button class="primary" type="button" onclick="startAccountWizard()">+ Connect account</button></div>
<div class="provider-grid">
<section class="provider-panel"><div class="provider-top"><span class="provider-icon google-icon">G</span><div><span class="mini-overline">GOOGLE</span><h3>Google Drive</h3><p>Use files from one or more Google accounts.</p></div><span class="provider-count">{{len .Accounts}} saved</span></div>
<div class="provider-list">{{range .Accounts}}<div class="provider-account"><div><strong>{{.Email}}</strong><small>Connected to Syncria</small></div><div class="provider-buttons"><button class="secondary" type="button" onclick="verifyConnectedAccount('{{.ID}}',this)">Test</button><form method="post" action="/action" class="inline" onsubmit="return confirm('Disconnect this Google Drive account? Your files will stay on Google Drive.')"><input type="hidden" name="action" value="remove_account"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="danger">Remove</button></form></div></div>{{else}}<div class="provider-empty">No Google Drive accounts connected yet.</div>{{end}}</div>
<button type="button" class="provider-add" onclick="startAccountWizard('google_drive')">+ Add Google Drive account <span aria-hidden="true">↗</span></button></section>
<section class="provider-panel"><div class="provider-top"><span class="provider-icon dropbox-icon">◆</span><div><span class="mini-overline">DROPBOX</span><h3>Dropbox</h3><p>Bring your Dropbox folders into a sync.</p></div><span class="provider-count">{{len .DropboxAccounts}} saved</span></div>
<div class="provider-list">{{range .DropboxAccounts}}<div class="provider-account"><div><strong>{{.Email}}</strong><small>Connected to Syncria</small></div><div class="provider-buttons"><form method="post" action="/action" class="inline" onsubmit="return confirm('Disconnect Dropbox? Your files stay in Dropbox.')"><input type="hidden" name="action" value="remove_dropbox_account"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="danger">Remove</button></form></div></div>{{else}}<div class="provider-empty">No Dropbox accounts connected yet.</div>{{end}}</div>
<button type="button" class="provider-add" onclick="startAccountWizard('dropbox')">+ Add Dropbox account <span aria-hidden="true">↗</span></button></section>
</div>
<section class="connection-drawer hidden" id="accountWizard"><div class="drawer-heading"><div><span class="mini-overline">NEW CLOUD CONNECTION</span><h3>Connect your cloud files</h3><p>Choose a service and sign in securely to connect it.</p></div><button class="secondary" type="button" onclick="$('accountWizard').classList.add('hidden')">Close ✕</button></div>
<div class="step-indicator"><span id="accountMarker1" class="active">1 · Choose service</span><span id="accountMarker2">2 · Sign in</span></div>
<div class="step active" id="accountStep1"><label class="field">Where are your cloud files?<select id="accountProvider" onchange="accountProviderChanged()"><option value="google_drive">Google Drive</option><option value="dropbox">Dropbox</option></select></label><p class="muted">This saves a connection, not your files. You'll choose folders when you create a sync.</p><div class="row"><button type="button" class="primary" onclick="showAccountStep(2)">Continue →</button><button type="button" class="secondary" onclick="navigate('settings')">Advanced sign-in settings</button></div></div>
<div class="step" id="accountStep2"><div id="googleAccountAuthorize"><p>Google Drive needs sign-in settings first. The person running Syncria can set them up in Settings.</p>
{{if .ClientID}}<a class="primary" href="/oauth/start">Sign in with Google Drive →</a>{{else}}<p class="wizard-status bad">Google sign-in is not set up yet.</p><button type="button" class="secondary" onclick="navigate('settings')">Open sign-in settings</button>{{end}}</div>
<div class="hidden" id="dropboxAccountAuthorize"><p>Dropbox needs sign-in settings first. The person running Syncria can set them up in Settings.</p>{{if .DropboxClientID}}<a class="primary" href="/oauth/dropbox/start">Sign in with Dropbox →</a>{{else}}<p class="wizard-status bad">Dropbox sign-in is not set up yet.</p><button type="button" class="secondary" onclick="navigate('settings')">Open sign-in settings</button>{{end}}</div><div class="row"><button type="button" class="secondary" onclick="showAccountStep(1)">← Back</button></div></div>
</section>
</section>
<section class="view" id="view-activity">
<div class="workspace-intro"><div><span class="mini-overline">THE LOG / REAL ACTIVITY</span><h2>See what's moving.</h2><p>Actual transfer counts and results from your syncs, updated as they run.</p></div><button class="secondary" type="button" onclick="refreshDashboardStatus()">Refresh activity ↻</button></div>
<div class="activity-scoreboard">
 <div class="activity-score activity-score-highlight"><small>CHECKS THAT WORKED</small><strong id="activitySuccessful">{{.Metrics.Runs}}</strong><span>completed checks</span></div>
 <div class="activity-score"><small>CHECKS WITH ERRORS</small><strong id="activityFailed">{{.Metrics.Failures}}</strong><span>need attention</span></div>
 <div class="activity-score"><small>FILES COPIED</small><strong id="activityFiles">—</strong><span>across all syncs</span></div>
 <div class="activity-score"><small>FILES SENT</small><strong id="activityUploaded">—</strong><span>to another place</span></div>
 <div class="activity-score"><small>FILES RECEIVED</small><strong id="activityDownloaded">—</strong><span>from another place</span></div>
 <div class="activity-score"><small>DATA TRANSFERRED</small><strong id="activityTransfer">—</strong><span>total</span></div>
</div>
<section class="activity-table-shell"><div class="section-heading"><div><span class="mini-overline">PER-SYNC RESULTS</span><h2>Sync details</h2><p>Numbers come from Syncria's actual transfer records.</p></div></div>
{{range .Jobs}}<div class="activity-row"><div class="activity-row-head"><div><strong>{{.Name}}</strong><small>{{if eq .Provider "bridge"}}{{if eq .CloudProvider "dropbox"}}Dropbox{{else}}Google Drive{{end}} ↔ NAS{{else if eq .Provider "nas_bridge"}}NAS ↔ NAS{{else}}Existing folder sync{{end}}</small></div><span class="activity-state" data-activity-running="{{.ID}}">Loading results…</span></div>
<div class="activity-row-numbers">
 <div><small>GOOD CHECKS</small><strong data-job-runs="{{.ID}}">{{.SuccessfulRuns}}</strong></div>
 <div><small>ERRORS</small><strong data-job-fails="{{.ID}}">{{.FailedRuns}}</strong></div>
 <div><small>FILES MOVED</small><strong data-job-files="{{.ID}}">—</strong></div>
 <div><small>BYTES MOVED</small><strong data-job-bytes="{{.ID}}">—</strong></div>
</div><p class="activity-row-detail" data-job-detail="{{.ID}}">Loading most recent result…</p></div>{{else}}<div class="flow-library-empty"><strong>No activity yet</strong><p>Create a sync to see its results here.</p></div>{{end}}</section>
</section>
<section class="view" id="view-updates">
<div class="workspace-intro"><div><span class="mini-overline">SYSTEM / RELEASES</span><h2>Always moving forward.</h2><p>Syncria installs updates from inside the app. See real download progress, verification, and restart status when an update runs.</p></div></div>
<div class="update-workspace">
<div>
<section class="panel release-panel"><div class="release-panel-heading"><span class="mini-overline">RELEASE CHANNEL / STABLE</span><span class="pill">BUILT-IN UPDATER</span></div>
<div class="release-well"><div><small>LATEST RELEASE</small><strong id="latestVersion">Checking GitHub Releases…</strong><p id="updateMessage">Looking for a newer version.</p></div><button class="secondary" type="button" onclick="checkUpdate(true)">Check again ↻</button></div>
<div class="update-revisions"><div><small>VERSION YOU'RE RUNNING</small><strong id="installedVersion">{{.Version}}</strong></div><div><small>VERSION AVAILABLE</small><strong id="updateAvailableVersion">Checking…</strong></div></div>
<div class="release-actions"><button id="installUpdateButton" class="primary" type="button" onclick="startUpdate('install')" disabled>Download & install update <span aria-hidden="true">→</span></button><button id="rollbackUpdateButton" class="secondary" type="button" onclick="startUpdate('rollback')">Restore previous version</button></div>
<p class="release-note">Installed here, in this Syncria instance. No separate updater, Docker image pull, or manual compilation.</p>
</section>
<section class="panel hidden" id="updateFlow" aria-live="polite" aria-label="Measured update progress"><div class="panel-head"><h2>Installing your update</h2><span class="pill" id="updateProgressOperation">Install</span></div>
<div class="update-progress-panel"><div class="update-progress-meta"><strong id="updatePhaseTitle">Preparing…</strong><strong id="updatePercent">—</strong></div>
<p id="updateProgressMessage" role="status">Waiting for update status…</p>
<progress id="updateProgressBar" class="update-meter" max="100" aria-label="Real download or verification progress"></progress>
<div class="update-connection" id="updateConnectionStatus">Waiting to hear from Syncria about your update.</div>
<div class="update-revisions"><div><small>FROM</small><strong id="updateFromVersion">—</strong></div><div><small>TO</small><strong id="updateToVersion">—</strong></div></div>
<div class="update-measures"><div><small>EXECUTABLE FILE</small><strong id="updateProgressAsset">—</strong></div><div><small id="updateTransferLabel">Actual downloaded bytes</small><strong id="updateProgressBytes">—</strong></div><div><small>BYTES VERIFIED</small><strong id="updateVerifiedBytes">—</strong></div><div><small>TIME TAKEN</small><strong id="updateElapsed">—</strong></div></div>
<div id="updateProgressTarget" class="muted"></div>
<ol class="update-steps" id="updateSteps"><li data-update-phase="checking">Check</li><li data-update-phase="manifest">Manifest</li><li data-update-phase="downloading">Download</li><li data-update-phase="verifying">Verify</li><li data-update-phase="staging">Install</li><li data-update-phase="restarting">Restart</li><li data-update-phase="complete">Ready</li></ol>
<p id="updateError" class="update-alert hidden" role="alert"></p>
<button id="updateReloadButton" type="button" class="primary hidden" onclick="window.location.assign('/?updated=1#updates')">Open updated Syncria →</button></div></section>
</div>
<div class="update-aside">
<section class="panel update-ready" id="updateReadyPanel"><div class="panel-head"><h2>How it works</h2><span class="quiet-label">3 STEPS</span></div><div class="update-explainer"><div><span>01</span><p><strong>Download</strong><small>The exact file arrives. You'll see real bytes downloaded.</small></p></div><div><span>02</span><p><strong>Verify</strong><small>Every byte is checked against the release.</small></p></div><div><span>03</span><p><strong>Restart</strong><small>Syncria only says success after the new version starts.</small></p></div></div></section>
<section class="panel hidden" id="updateHistoryPanel"><div class="panel-head"><h2>Live update log</h2><span class="status" id="updateLogState">Waiting</span></div><p class="muted">Real events recorded by Syncria, with timestamps.</p><ol class="update-timeline" id="updateEventTimeline" aria-live="polite"><li><strong>Ready</strong><small>Events appear here during an update.</small></li></ol></section>
<section class="panel"><details class="advanced-update"><summary>Install from files (advanced)</summary><div class="advanced-update-body"><h2>Manual installation</h2><p class="muted">Use this only when automatic downloads are unavailable. Supply a matching app file and SHA256SUMS. Your previous version remains available for rollback.</p><form id="manualUpdateForm" action="/api/update/manual" method="post" enctype="multipart/form-data" onsubmit="return submitManualUpdate(event)"><label class="field">App executable<input type="file" name="binary" id="manualBinary" required></label><label class="field">SHA256SUMS<input type="file" name="checksums" required></label><button id="manualUpdateButton" class="secondary" type="submit">Verify & install files</button></form></div></details></section>
</div></div>
</section>
<section class="view" id="view-settings"><div class="workspace-intro"><div><span class="mini-overline">SYSTEM / CONNECTION SETTINGS</span><h2>Fine-tune Syncria.</h2><p>Advanced connection settings live here, away from everyday syncing. You usually only set these once.</p></div></div><section class="panel"><h2>Google sign-in settings (advanced)</h2><p>For an administrator: create a Google Cloud web sign-in app, enable the Google Drive API, and add the address shown below as its redirect URL:</p><div class="selected"><code>{{.Callback}}</code></div><form method="post" action="/action"><input type="hidden" name="action" value="oauth_settings"><label class="field">Google app ID (Client ID)<input name="client_id" value="{{.ClientID}}" required></label><label class="field">Google app secret<input name="client_secret" type="password" placeholder="Leave blank to keep existing secret"></label><label class="field">Your Syncria web address<input name="base_url" value="{{.BaseURL}}" required></label><button class="primary">Save Google sign-in settings</button></form></section><section class="panel"><h2>Dropbox sign-in settings (advanced)</h2>
<p>For an administrator: add this return address to your Dropbox developer app:</p><div class="selected"><code>{{.DropboxCallback}}</code></div>
<form method="post" action="/action"><input type="hidden" name="action" value="dropbox_oauth_settings">
<label class="field">Dropbox app key<input name="client_id" value="{{.DropboxClientID}}" required></label>
<label class="field">Dropbox app secret<input name="client_secret" type="password" placeholder="Leave empty to keep existing secret"></label>
<button type="submit" class="primary">Save Dropbox sign-in settings</button></form></section></section>
{{end}}</div></main>
<div class="picker" id="picker" role="dialog" aria-modal="true" aria-labelledby="pickerTitle"><div class="picker-box"><div class="picker-head"><div class="panel-head" style="margin:0"><h2 id="pickerTitle">Select folder</h2><button class="secondary" type="button" onclick="closePicker()">✕ Close</button></div><div id="pickerPath" class="picker-path">Loading…</div></div><div class="row picker-actions"><button class="secondary" type="button" id="pickerUp" onclick="pickerUp()">← Back</button><button class="secondary" type="button" id="pickerNew" onclick="createFolder()">+ New folder</button></div><div id="folderList" class="folder-list" aria-live="polite"></div><div class="picker-footer"><small id="pickerStatus" class="muted">Choose a folder.</small><button class="primary" type="button" id="pickerSelect" onclick="selectFolder()">Select folder</button></div></div></div>
<script>
const $=id=>document.getElementById(id);
let currentView='overview',pickerType='',folder='',parent='',folderName='',pickSeq=0;
function navigate(view){
 if(!$('view-'+view))return;
 currentView=view;
 document.querySelectorAll('.view').forEach(el=>el.classList.toggle('active',el.id==='view-'+view));
 document.querySelectorAll('.nav button').forEach(el=>{
  const active=el.dataset.view===view;
  el.classList.toggle('active',active);
  if(active)el.setAttribute('aria-current','page');else el.removeAttribute('aria-current');
 });
 const pages={
  overview:['Your overview','OVERVIEW'],
  sync:['Your syncs','MY SYNCS'],
  storage:['Storage devices','MY NAS BOXES'],
  accounts:['Cloud connections','CLOUD ACCOUNTS'],
  activity:['Transfer activity','ACTIVITY'],
  updates:['App updates','UPDATES'],
  settings:['Settings','PREFERENCES']
 };
 const page=pages[view]||[view,view.toUpperCase()];
 $('pageTitle').textContent=page[0];
 const breadcrumb=$('pageBreadcrumb');if(breadcrumb)breadcrumb.textContent=page[1];
 history.replaceState(null,'','#'+view);
 if(view==='updates')checkUpdate();
}
document.querySelectorAll('.nav button').forEach(b=>b.addEventListener('click',()=>navigate(b.dataset.view)));
let syncReviewed=false,syncStep=1,selectedNASTargetID='';
function showSyncStep(step){
 syncStep=step;if(step!==4)syncReviewed=false;
 for(let i=1;i<=4;i++){$('wizard'+i).classList.toggle('active',i===step);$('stepMarker'+i).classList.toggle('active',i===step);}
 $('wizard').scrollIntoView({behavior:'smooth',block:'start'});
}
function setWizardStatus(message,bad=false){
 const el=$('syncWizardStatus');if(!el)return;
 el.textContent=message;el.classList.toggle('bad',bad);el.classList.toggle('good',!bad);
}
function startWizard(){
 navigate('sync');$('syncWizard').reset();
 $('createSyncButton').disabled=false;$('wizard').classList.remove('hidden');
 $('remoteValue').value='root';$('selectedDriveLabel').textContent='Entire My Drive';
 $('nasSourceFolderValue').value='.';$('selectedNASSourceLabel').textContent='Share root';
 document.querySelectorAll('#nasTargets input[type=hidden]').forEach(el=>el.value='.');
 document.querySelectorAll('#nasTargets small[id^=nas-label-]').forEach(el=>el.textContent='Share root');
 sourceProviderChanged();relationshipTypeChanged();scheduleModeChanged();nasDestChanged();showSyncStep(1);
}
function closeSyncWizard(){$('wizard').classList.add('hidden');syncReviewed=false;}
function relationshipTypeChanged(){
 syncReviewed=false;
 const cloud=$('relationshipType').value==='cloud';
 $('cloudSourceFields').classList.toggle('hidden',!cloud);
 $('nasSourceFields').classList.toggle('hidden',cloud);
 $('nasDirectionFields').classList.toggle('hidden',cloud);
 $('crossNASField').classList.toggle('hidden',!cloud);
 $('syncAccount').required=cloud&&$('sourceProvider').value==='google_drive';
 $('syncDropboxAccount').required=cloud&&$('sourceProvider').value==='dropbox';
 $('nasSourceId').required=!cloud;
 nasDirectionChanged();
 setWizardStatus(cloud?'Choose the account and folder where your files are now.':'Choose the first NAS box and the folder that already has your files.');
}
function sourceProviderChanged(){
 syncReviewed=false;
 const google=$('sourceProvider').value==='google_drive';
 $('googleSourceFields').classList.toggle('hidden',!google);
 $('dropboxSourceFields').classList.toggle('hidden',google);
 $('syncAccount').required=google&&$('relationshipType').value==='cloud';
 $('syncDropboxAccount').required=!google&&$('relationshipType').value==='cloud';
 resetDrive();
}
function resetDrive(){
 $('remoteValue').value=$('sourceProvider').value==='dropbox'?'':'root';
 $('selectedDriveLabel').textContent=$('sourceProvider').value==='dropbox'?'Entire Dropbox':'Entire My Drive';
 syncReviewed=false;
}
function nasSourceChanged(){
 $('nasSourceFolderValue').value='.';$('selectedNASSourceLabel').textContent='Share root';syncReviewed=false;
}
function nasDestChanged(){
 syncReviewed=false;
 const chosen=selectedNASDestinations();
 $('nasDestStatus').textContent=chosen.length?chosen.length+' NAS box(es) chosen. Each box can have a different folder.':'Choose at least one NAS box to receive copies.';
}
function selectedNASDestinations(){
 return [...document.querySelectorAll('#nasTargets input[name=nas_ids]:checked')].map(el=>({id:el.value,
  folder:$('nas-folder-'+el.value)?.value||'.',
  label:el.closest('.nas-target')?.querySelector('strong')?.textContent||el.value
 }));
}
function browseNASDestination(id){
 const checkbox=[...document.querySelectorAll('#nasTargets input[name=nas_ids]')].find(el=>el.value===id);
 if(checkbox&&!checkbox.checked){checkbox.checked=true;nasDestChanged();}
 selectedNASTargetID=id;openPicker('nas-target');
}
function scheduleModeChanged(){
 const changed=$('scheduleMode').value==='changes';
 $('syncInterval').closest('.field').classList.toggle('hidden',changed);
 if(changed)$('syncInterval').value='5';
}
function nasDirectionChanged(){
 const nas=$('relationshipType').value==='nas',bi=$('nasDirection').value==='bidirectional';
 $('nasConflictFields').classList.toggle('hidden',!(nas&&bi));
}
async function verifyWizard(kind,data){
 const body=new URLSearchParams({kind,...data});
 const res=await fetch('/api/wizard/verify',{method:'POST',body,credentials:'same-origin',cache:'no-store'});
 let result;try{result=await res.json();}catch(_){throw Error('Source verification failed (HTTP '+res.status+')');}
 if(!res.ok||!result.ok)throw Error(result.message||'Source unavailable');
 return result.message;
}
async function verifyNAS(id,pathValue){
 const response=await fetch('/api/browse/nas?id='+encodeURIComponent(id)+'&path='+encodeURIComponent(pathValue||'.'),{credentials:'same-origin',cache:'no-store'});
 if(!response.ok)throw Error('NAS '+id+': '+(await response.text()).trim());
 return response.json();
}
async function verifySource(){
 if($('relationshipType').value==='nas'){
  if(!$('nasSourceId').reportValidity())throw Error('Choose the NAS box that already has your files.');
  await verifyNAS($('nasSourceId').value,$('nasSourceFolderValue').value);
  return 'Good! Syncria can open your first NAS folder.';
 }
 const provider=$('sourceProvider').value;
 if(provider==='dropbox'){
  if(!$('syncDropboxAccount').reportValidity())throw Error('Choose the Dropbox account that has your files.');
  const url='/api/browse/dropbox?account='+encodeURIComponent($('syncDropboxAccount').value)+'&path='+encodeURIComponent($('remoteValue').value);
  const resp=await fetch(url,{credentials:'same-origin',cache:'no-store'});
  if(!resp.ok)throw Error((await resp.text()).trim());
  return 'Good! Syncria can open this Dropbox folder.';
 }
 if(!$('syncAccount').reportValidity())throw Error('Choose the Google Drive account that has your files.');
 return verifyWizard('drive',{account:$('syncAccount').value,folder:$('remoteValue').value});
}
async function wizardNext(){
 if(!$('syncName').reportValidity())return;
 const btn=$('verifySourceButton');btn.disabled=true;setWizardStatus('Checking that Syncria can open this folder…');
 try{setWizardStatus(await verifySource());showSyncStep(2);}
 catch(e){setWizardStatus(e.message,true);}
 finally{btn.disabled=false;}
}
async function wizardReview(){
 const btn=$('verifyTargetButton');btn.disabled=true;
 try{
  const targets=selectedNASDestinations();
  if(targets.length<1)throw Error('Choose at least one NAS box for the copies.');
  if($('relationshipType').value==='nas'){
   const source=$('nasSourceId').value,pathValue=$('nasSourceFolderValue').value;
   for(const target of targets){
    const p=pathValue.replace(/\/$/,'')||'.',q=target.folder.replace(/\/$/,'')||'.';
    if(source===target.id&&(p===q||p==='.'||q==='.'||p.startsWith(q+'/')||q.startsWith(p+'/')))throw Error('The same folder cannot be both the starting folder and a copy. Pick another folder on '+target.label+'.');
   }
  }
  for(const target of targets)await verifyNAS(target.id,target.folder);
  await verifySource();
  $('nasDestStatus').textContent='Good! Syncria can open '+targets.length+' chosen NAS folder'+(targets.length===1?'':'s')+'.';
  showSyncStep(3);
 }catch(e){$('nasDestStatus').textContent=e.message;alert('Syncria could not open a folder yet. Check your folder and sign-in details, then try again.\n\nDetails: '+e.message);}
 finally{btn.disabled=false;}
}
function wizardFinalize(){
 if(syncStep!==3)return;
 const nas=$('relationshipType').value==='nas';
 $('reviewName').textContent=$('syncName').value.trim();
 $('reviewProvider').textContent=nas?'NAS':$('sourceProvider').selectedOptions[0]?.textContent;
 $('reviewSource').textContent=nas?$('nasSourceId').selectedOptions[0]?.textContent+' / '+$('selectedNASSourceLabel').textContent:
  ($('sourceProvider').value==='dropbox' ? $('syncDropboxAccount').selectedOptions[0]?.textContent:$('syncAccount').selectedOptions[0]?.textContent)+' / '+$('selectedDriveLabel').textContent;
 $('reviewTarget').textContent=selectedNASDestinations().map(x=>x.label+' ('+x.folder+')').join(', ');
 $('reviewInterval').textContent=$('scheduleMode').value==='changes'?'When files change, plus checks every 5 minutes':$('syncInterval').selectedOptions[0]?.textContent||'';
 $('reviewDirection').textContent=nas?($('nasDirection').value==='one_way'?'One way: first NAS to the others':'Both ways · '+$('conflictPolicy').selectedOptions[0]?.textContent):
  ('Cloud and NAS'+($('crossNASField').querySelector('input').checked?' · Copy changes between NAS boxes through cloud':' · No cross-NAS copying through cloud'));
 $('reviewDeletion').textContent=($('propagateDeletions').checked?'Delete matching copies too':'Keep other copies if I delete a file')+
  ($('recoveryEnabled').checked?' · 10-minute Undo on':' · No Undo window');
 showSyncStep(4);syncReviewed=true;
}
function submitSyncWizard(event){
 event.preventDefault();if(syncStep!==4||!syncReviewed)return false;
 syncReviewed=false;$('createSyncButton').disabled=true;$('syncWizard').submit();return false;
}
function toggleDirectNAS(force){
 navigate('storage');
 const panel=$('directNASPanel'),open=force===undefined?panel.classList.contains('hidden'):!!force;
 panel.classList.toggle('hidden',!open);
 if(open)panel.scrollIntoView({behavior:'smooth',block:'start'});
}
function startAccountWizard(provider){
 navigate('accounts');$('accountWizard').classList.remove('hidden');
 if(provider)$('accountProvider').value=provider;
 accountProviderChanged();showAccountStep(1);$('accountWizard').scrollIntoView({behavior:'smooth'});
}
function accountProviderChanged(){
 const dropbox=$('accountProvider').value==='dropbox';
 $('googleAccountAuthorize').classList.toggle('hidden',dropbox);
 $('dropboxAccountAuthorize').classList.toggle('hidden',!dropbox);
}
function showAccountStep(step){
 for(let i=1;i<=2;i++){$('accountStep'+i).classList.toggle('active',i===step);$('accountMarker'+i).classList.toggle('active',i===step);}
 accountProviderChanged();
}
async function verifyConnectedAccount(id,button){
 button.disabled=true;button.textContent='Verifying…';
 try{const result=await verifyWizard('drive',{account:id,folder:'root'});button.textContent='Verified';button.title=result;}
 catch(e){button.textContent='Could not connect';button.title='Check the account sign-in or try again. Details: '+e.message;}
 finally{button.disabled=false;}
}
let updatePollingBusy=false,updateSubmitting=false,updateRunning=false,updateInstallAvailable=false;
let lastUpdatePhase='idle',lastUpdateLog='',lastConfirmedUpdateStatus='';
let manualUploadInProgress=false;
let completedUpdateTimer=null,hiddenCompletedUpdate='';
function resetUpdateView(){
 updateRunning=false;
 lastUpdatePhase='idle';
 lastUpdateLog='';
 const flow=$('updateFlow'),history=$('updateHistoryPanel'),ready=$('updateReadyPanel');
 if(flow)flow.classList.add('hidden');
 if(history)history.classList.add('hidden');
 if(ready)ready.classList.remove('hidden');
 const banner=$('updateLiveBanner');if(banner)banner.classList.add('hidden');
 const logState=$('updateLogState');if(logState)logState.textContent='Ready';
 const timeline=$('updateEventTimeline');
 if(timeline)timeline.replaceChildren();
 const install=$('installUpdateButton'),rollback=$('rollbackUpdateButton');
 if(install)install.disabled=updateSubmitting||!updateInstallAvailable;
 if(rollback)rollback.disabled=updateSubmitting;
}

const updatePhaseTitles={
 checking:'Checking published release',manifest:'Reading release checksums',
 receiving:'Receiving manual update',downloading:'Downloading exact executable',
 verifying:'Verifying SHA-256 bytes',staging:'Installing verified executable',
 restarting:'Restart requested',complete:'Installed and verified',failed:'Update failed',idle:'No update'
};
const updateSequence=['checking','manifest','receiving','downloading','verifying','staging','restarting','complete'];
function updateClock(value){
 if(!value)return 'Time unavailable';
 const date=new Date(value);
 return Number.isFinite(date.getTime())?date.toLocaleString():'Time unavailable';
}
function renderUpdateEvents(events,phase){
 if(!$('updateEventTimeline'))return;
 const serialized=JSON.stringify(events||[]);
 if(serialized===lastUpdateLog)return;
 lastUpdateLog=serialized;
 const list=$('updateEventTimeline');list.replaceChildren();
 if(!Array.isArray(events)||!events.length){
  const li=document.createElement('li');
  const strong=document.createElement('strong');strong.textContent='No recorded events';
  li.appendChild(strong);list.appendChild(li);return;
 }
 for(const [i,event] of events.entries()){
  const li=document.createElement('li');
  const final=i===events.length-1;
  li.classList.toggle('active',final&&phase!=='failed');
  li.classList.toggle('failed',final&&phase==='failed');
  const strong=document.createElement('strong');strong.textContent=event.detail||event.phase||'Update event';
  const small=document.createElement('small');small.textContent=updateClock(event.at);
  li.append(strong,small);list.appendChild(li);
 }
}
function setUpdateProgressBar(phase,current,total){
 const bar=$('updateProgressBar'),label=$('updatePercent');
 if(!bar||!label)return;
 if(phase==='complete'){bar.value=100;label.textContent='Completed';return;}
 if((phase==='downloading'||phase==='verifying'||phase==='receiving')&&total>0){
  const pc=Math.min(100,Math.max(0,Math.floor(current*100/total)));
  bar.value=pc;label.textContent=pc+'% of '+(phase==='verifying'?'SHA-256 hash':phase==='receiving'?'upload':'download');
 }else{
  bar.removeAttribute('value');
  label.textContent=phase==='failed'?'Failed':'Measured progress only';
 }
}
function renderUpdateProgress(p){
 if(!p||!p.phase)return;
 const phase=p.phase;
 if(phase==='idle'){resetUpdateView();return;}
 const completionKey=phase==='complete'?((p.started_at||'')+'|'+(p.target||'')+'|'+(p.updated_at||'')):'';
 if(phase==='complete'&&completionKey===hiddenCompletedUpdate){resetUpdateView();return;}
 if(phase!=='complete'&&completedUpdateTimer){
  clearTimeout(completedUpdateTimer);completedUpdateTimer=null;
 }
 lastUpdatePhase=phase;
 updateRunning=!!p.active;
 $('updateFlow')?.classList.remove('hidden');
 $('updateHistoryPanel')?.classList.remove('hidden');
 $('updateReadyPanel')?.classList.add('hidden');
 if(phase==='complete'&&!completedUpdateTimer){
  // A success is a short acknowledgement, not a permanent activity log.
  completedUpdateTimer=setTimeout(()=>{
   completedUpdateTimer=null;
   hiddenCompletedUpdate=completionKey;
   resetUpdateView();
   checkUpdate();
  },3500);
 }
 const events=Array.isArray(p.events)?p.events:[];
 renderUpdateEvents(events,phase);
 const logState=$('updateLogState');
 if(logState)logState.textContent=phase==='complete'?'Verified':phase==='failed'?'Failed':p.active?'Recording live':'Last update';
 const set=(id,value)=>{const el=$(id);if(el)el.textContent=value;};
 set('updatePhaseTitle',updatePhaseTitles[phase]||phase);
 set('updateProgressMessage',p.message||'No stage description provided by the server');
 set('updateProgressOperation',p.operation==='rollback'?'Rollback':p.operation==='manual'?'Manual update':'Release install');
 set('updateFromVersion',p.from||p.current||'—');
 set('updateToVersion',p.target||'Awaiting release metadata');
 set('updateProgressAsset',p.asset||'Awaiting asset selection');
 set('updateProgressTarget',p.target?'Target release: '+p.target:'');
 const bytes=Number(p.downloaded||0),total=Number(p.total||0);
 const verified=Number(p.verified_bytes||0),verificationTotal=Number(p.verified_total||0);
 set('updateTransferLabel',p.operation==='manual'?'Actual uploaded bytes':'Actual downloaded bytes');
 if(!(manualUploadInProgress&&phase==='receiving')){
  set('updateProgressBytes',bytes>0||total>0?formatBytes(bytes)+(total>0?' / '+formatBytes(total):' · size not reported'):'No binary bytes received yet');
 }
 set('updateVerifiedBytes',verified>0||verificationTotal>0?formatBytes(verified)+(verificationTotal>0?' / '+formatBytes(verificationTotal):''):'Not started');
 const secs=Math.max(0,Number(p.elapsed_seconds||0));
 set('updateElapsed',Math.floor(secs/60)+'m '+Math.floor(secs%60)+'s');
 if(!(manualUploadInProgress&&phase==='receiving')){setUpdateProgressBar(phase,phase==='verifying'?verified:bytes,phase==='verifying'?verificationTotal:total);}
 const error=$('updateError');
 if(error){
  error.classList.toggle('hidden',phase!=='failed');
  error.textContent=phase==='failed'?(p.error||'The updater recorded a failure without a reason'):'';
 }
 const reconnect=$('updateConnectionStatus');
 if(reconnect){
  reconnect.classList.remove('lost');
  reconnect.textContent='Connected · status verified at '+updateClock(p.updated_at)+(p.current?' · running '+p.current:'');
 }
 const button=$('updateReloadButton');
 if(button)button.classList.toggle('hidden',phase!=='complete'||(p.operation!=='manual'&&(!p.target||p.current!==p.target)));
 for(const li of document.querySelectorAll('[data-update-phase]')){
  const index=updateSequence.indexOf(li.dataset.updatePhase);
  const phaseIndex=updateSequence.indexOf(phase);
  li.classList.toggle('active',li.dataset.updatePhase===phase);
  li.classList.toggle('done',phase==='complete'||(index>=0&&phaseIndex>=0&&index<phaseIndex));
  li.classList.toggle('failed',phase==='failed'&&li.dataset.updatePhase==='restarting');
 }
 if(p.current){
  set('installedVersion',p.current);set('sidebarVersion',p.current);
 }
 const install=$('installUpdateButton'),rollback=$('rollbackUpdateButton');
 if(install)install.disabled=updateRunning||updateSubmitting||!updateInstallAvailable;
 if(rollback)rollback.disabled=updateRunning||updateSubmitting;
 const banner=$('updateLiveBanner'),bannerText=$('updateLiveBannerStatus');
 if(banner){
  banner.classList.toggle('hidden',!updateRunning||currentView==='updates');
  if(bannerText)bannerText.textContent=(p.target||'Release')+' · '+(p.message||phase);
 }
}
async function pollUpdateProgress(){
 if(updatePollingBusy||!$('updateFlow'))return;
 if(!updateRunning&&currentView!=='updates'&&lastUpdatePhase==='idle')return;
 updatePollingBusy=true;
 try{
  const response=await fetch('/api/update/progress',{credentials:'same-origin',cache:'no-store'});
  if(response.status===401){throw Error('Authentication expired (401)');}
  if(!response.ok)throw Error('HTTP '+response.status);
  const payload=await response.json();
  lastConfirmedUpdateStatus=payload.updated_at||'';
  renderUpdateProgress(payload);
 }catch(error){
  if(updateRunning||currentView==='updates'){
   const reconnect=$('updateConnectionStatus');
   if(reconnect){
    reconnect.classList.add('lost');
    reconnect.textContent='Status endpoint unreachable: '+error.message+'. Last recorded step: '+(updatePhaseTitles[lastUpdatePhase]||lastUpdatePhase)+'. The browser cannot confirm completion until the new Syncria process responds.';
   }
  }
 }finally{updatePollingBusy=false;}
}
async function startUpdate(operation){
 if(updateSubmitting||updateRunning)return;
 hiddenCompletedUpdate='';
 if(completedUpdateTimer){clearTimeout(completedUpdateTimer);completedUpdateTimer=null;}
 if(!confirm(operation==='rollback'?'Restore Syncria\'s saved previous executable?':'Download, verify and install the latest published Syncria release?'))return;
 navigate('updates');
 $('updateFlow').classList.remove('hidden');
 updateSubmitting=true;
 const set=(id,value)=>{const el=$(id);if(el)el.textContent=value;};
 set('updatePhaseTitle','Requesting update');
 set('updateProgressMessage','Submitting update request to this Syncria instance.');
 set('updateConnectionStatus','Connecting to the update status endpoint…');
 $('updateError').classList.add('hidden');$('updateReloadButton').classList.add('hidden');
 $('installUpdateButton').disabled=true;$('rollbackUpdateButton').disabled=true;
 try{
  const response=await fetch('/api/update/progress',{method:'POST',credentials:'same-origin',cache:'no-store',body:new URLSearchParams({operation})});
  if(!response.ok)throw Error((await response.text()).trim()||('HTTP '+response.status));
  renderUpdateProgress(await response.json());
  updateRunning=true;
 }catch(error){
  updateRunning=false;
  set('updatePhaseTitle','Update request rejected');
  set('updateProgressMessage','Updater did not accept this request. No new release has been installed.');
  set('updateError',error.message);$('updateError').classList.remove('hidden');
 }finally{updateSubmitting=false;await pollUpdateProgress();}
}
async function checkUpdate(force=false){
 if(!$('latestVersion'))return;
 $('latestVersion').textContent='Checking GitHub Releases…';
 $('updateMessage').textContent='';updateInstallAvailable=false;
 $('installUpdateButton').disabled=true;
 try{
  const res=await fetch('/api/update/status'+(force?'?refresh=1':''),{credentials:'same-origin',cache:'no-store'});
  if(!res.ok)throw Error('Release check HTTP '+res.status);
  const result=await res.json();
  if(result.error)throw Error(result.error);
  $('latestVersion').textContent=result.latest||'No published release';
  $('updateAvailableVersion').textContent=result.latest||'None';
  $('updateMessage').textContent=result.pending?result.message||'Awaiting release binaries':result.available?'Verified release assets available for this platform':'Installed release is current';
  updateInstallAvailable=!!result.available;
  $('installUpdateButton').disabled=updateRunning||updateSubmitting||!updateInstallAvailable;
 }catch(error){
  $('latestVersion').textContent='Release check failed';
  $('updateAvailableVersion').textContent='Unavailable';
  $('updateMessage').textContent=error.message;
 }
 await pollUpdateProgress();
}
async function submitManualUpdate(event){
 event.preventDefault();
 hiddenCompletedUpdate='';
 if(completedUpdateTimer){clearTimeout(completedUpdateTimer);completedUpdateTimer=null;}
 if(updateSubmitting||updateRunning)return false;
 if(!confirm('Install the selected executable after verifying its SHA256SUMS?'))return false;
 navigate('updates');
 const form=event.currentTarget,button=$('manualUpdateButton');
 updateSubmitting=true;manualUploadInProgress=true;button.disabled=true;
 $('updateFlow').classList.remove('hidden');
 $('updatePhaseTitle').textContent='Submitting manual update';
 $('updateProgressMessage').textContent='Uploading the selected executable to Syncria.';
 try{
  const request=new XMLHttpRequest();
  const result=await new Promise((resolve,reject)=>{
   request.open('POST','/api/update/manual',true);
   request.responseType='text';
   request.upload.onprogress=e=>{
    if(e.lengthComputable)setUpdateProgressBar('receiving',e.loaded,e.total);
    $('updateProgressBytes').textContent=formatBytes(e.loaded)+(e.lengthComputable?' / '+formatBytes(e.total):'');
   };
   request.onload=()=>request.status>=200&&request.status<300?resolve(request.responseText):reject(Error(request.responseText||'Manual install HTTP '+request.status));
   request.onerror=()=>reject(Error('Manual update upload connection failed'));
   request.send(new FormData(form));
  });
  await pollUpdateProgress();
 }catch(error){
  $('updatePhaseTitle').textContent='Manual update request failed';
  $('updateError').textContent=error.message;$('updateError').classList.remove('hidden');
 }finally{manualUploadInProgress=false;updateSubmitting=false;button.disabled=false;await pollUpdateProgress();}
 return false;
}
setInterval(pollUpdateProgress,1000);
function openPicker(type){
 const nas=type==='nas-source'||type==='nas-target';
 if(!nas&&type!=='drive'&&type!=='dropbox')return;
 if(nas){
  const id=type==='nas-source' ? $('nasSourceId').value:selectedNASTargetID;
  if(!id){alert('Choose a NAS connection before browsing folders.');return;}
 }
 pickerType=type;folder=null;parent=null;folderName='';
 $('picker').classList.add('open');
 $('pickerTitle').textContent=type==='drive'?'Browse Google Drive':type==='dropbox'?'Browse Dropbox':'Browse NAS folders';
 $('pickerNew').classList.toggle('hidden',!nas);
 const initial=type==='drive'?'root':type==='dropbox'?'':type==='nas-source'?$('nasSourceFolderValue').value:$('nas-folder-'+selectedNASTargetID).value;
 loadFolder(initial);
}
function closePicker(){pickSeq++;$('picker').classList.remove('open');}
async function loadFolder(selected,label){
 const sequence=++pickSeq;
 $('pickerStatus').textContent='Listing remote folders…';$('pickerSelect').disabled=true;
 $('folderList').replaceChildren();
 const q=encodeURIComponent(selected??'');
 let endpoint='';
 if(pickerType==='drive')endpoint='/api/browse/drive?account='+encodeURIComponent($('syncAccount').value)+'&parent='+q;
 else if(pickerType==='dropbox')endpoint='/api/browse/dropbox?account='+encodeURIComponent($('syncDropboxAccount').value)+'&path='+q;
 else{
  const id=pickerType==='nas-source'?$('nasSourceId').value:selectedNASTargetID;
  endpoint='/api/browse/nas?id='+encodeURIComponent(id)+'&path='+q;
 }
 try{
  const response=await fetch(endpoint,{credentials:'same-origin',cache:'no-store'});
  if(!response.ok)throw Error((await response.text()).trim());
  const data=await response.json();if(sequence!==pickSeq)return;
  folder=data.path;parent=data.parent===undefined?null:data.parent;
  folderName=label||(pickerType==='drive'?(folder==='root'?'Entire My Drive':folder):
   pickerType==='dropbox'?(folder===''?'Entire Dropbox':folder):(folder==='.'?'Share root':folder));
  $('pickerPath').textContent=folderName;
  $('pickerUp').disabled=parent===null;
  $('pickerSelect').disabled=false;
  for(const item of data.entries||[]){
   const button=document.createElement('button');
   button.className='folder-row';button.type='button';button.textContent='📁 '+item.name;
   button.addEventListener('click',()=>loadFolder(item.id,item.name));
   $('folderList').appendChild(button);
  }
  $('pickerStatus').textContent=data.message||((data.entries||[]).length+' folders. Select this folder or browse deeper.');
 }catch(e){
  if(sequence!==pickSeq)return;
  $('pickerStatus').textContent='Folder listing failed: '+e.message;
 }
}
function pickerUp(){if(parent!==null)loadFolder(parent);}
function selectFolder(){
 if(folder===null||!pickerType)return;
 syncReviewed=false;
 if(pickerType==='drive'||pickerType==='dropbox'){
  $('remoteValue').value=folder;$('selectedDriveLabel').textContent=folderName||folder||'Entire Dropbox';
 }else if(pickerType==='nas-source'){
  $('nasSourceFolderValue').value=folder;$('selectedNASSourceLabel').textContent=folderName||'Share root';
 }else if(pickerType==='nas-target'){
  $('nas-folder-'+selectedNASTargetID).value=folder;
  $('nas-label-'+selectedNASTargetID).textContent=folderName||'Share root';
 }
 closePicker();
}
async function createFolder(){
 if((pickerType!=='nas-source'&&pickerType!=='nas-target')||folder===null)return;
 let name=prompt('New folder on NAS');if(name===null)return;
 name=name.trim();if(!name)return;
 const id=pickerType==='nas-source'?$('nasSourceId').value:selectedNASTargetID;
 try{
  const body=new URLSearchParams({id,parent:folder,name});
  const response=await fetch('/api/browse/nas/create',{method:'POST',body,credentials:'same-origin'});
  if(!response.ok)throw Error((await response.text()).trim());
  await loadFolder(folder);
 }catch(e){$('pickerStatus').textContent='NAS folder creation failed: '+e.message;}
}
// Never navigate away from an active UI session in response to background polling.
// The Updates screen checks the installed version on demand. The server restart
// during an explicit installation is handled by the existing update workflow.
// Poll only small status fields; never replace forms, panels, or the document.
function formatBytes(value){
  let n=Number(value)||0;
  if(n<1024)return n.toLocaleString()+' B';
  const units=['KB','MB','GB','TB','PB'];let i=-1;
  do{n/=1024;i++;}while(n>=1024&&i<units.length-1);
  return n.toFixed(n<10?1:0)+' '+units[i];
}
let dashboardPollBusy=false,dashboardSessionExpired=false;
async function refreshDashboardStatus(){
 if(document.hidden||dashboardPollBusy||dashboardSessionExpired||!$('view-overview'))return;
 dashboardPollBusy=true;
 try{
  const res=await fetch('/api/dashboard/status',{cache:'no-store',credentials:'same-origin'});
  if(res.status===401){dashboardSessionExpired=true;const banner=$('sessionExpiredNotice');if(banner)banner.classList.remove('hidden');return;}
  if(!res.ok)return;
  const data=await res.json();
  if($('installedVersion')&&data.current)$('installedVersion').textContent=data.current;
  if($('sidebarVersion')&&data.current)$('sidebarVersion').textContent=data.current;
   if(data.metrics){
     const m=data.metrics;
     const set=(id,value)=>{const el=$(id);if(el)el.textContent=value;};
     const num=n=>Number(n||0).toLocaleString();
     set('metricRuns',num(m.Runs));
     set('metricFailures',num(m.Failures));
     set('metricFiles',num(m.TrackedFiles));
     set('metricBandwidth',formatBytes(Number(m.UploadBytes||0)+Number(m.DownloadBytes||0)));
     set('activitySuccessful',num((m.Runs||0)-(m.Failures||0)));
     set('activityFailed',num(m.Failures));
     set('activityFiles',num((m.Uploaded||0)+(m.Downloaded||0)));
     set('activityUploaded',formatBytes(m.UploadBytes));
     set('activityDownloaded',formatBytes(m.DownloadBytes));
     set('activityTransfer',formatBytes(Number(m.UploadBytes||0)+Number(m.DownloadBytes||0)));
   }
  for(const job of data.jobs||[]){
   for(const el of document.querySelectorAll('[data-job-enabled]')){
    if(el.dataset.jobEnabled!==job.id)continue;
    const on=!!(job.enabled||job.running);
    el.classList.toggle('is-on',on);
    el.classList.toggle('is-off',!on);
    let dot=el.querySelector('span');
    if(!dot){dot=document.createElement('span');el.replaceChildren(dot);}
    const label=job.running?' Syncing':job.enabled?' On':' Paused';
    if(el.lastChild?.nodeType===Node.TEXT_NODE){el.lastChild.textContent=label;}
    else{el.appendChild(document.createTextNode(label));}
   }
   const run=document.querySelector('[data-job-run="'+CSS.escape(job.id)+'"]');
   if(run)run.textContent=job.last_run==='Never run'?'Never run':'Last run '+job.last_run;
   const summary=document.querySelector('[data-job-summary="'+CSS.escape(job.id)+'"]');
   if(summary)summary.textContent=job.summary||'';
   const error=document.querySelector('[data-job-error="'+CSS.escape(job.id)+'"]');
   if(error){error.hidden=!job.error;error.textContent=job.error?'Error: '+job.error:'';}
    const el=(attr)=>document.querySelector('['+attr+'="'+CSS.escape(job.id)+'"]');
    const bytes=Number(job.total_upload_bytes||0)+Number(job.total_download_bytes||0);
    const files=Number(job.total_uploads||0)+Number(job.total_downloads||0);
    const line=el('data-job-statline');
    if(line)line.textContent=files.toLocaleString()+' files moved · '+formatBytes(bytes)+' transferred · '+job.successful_runs+' successful / '+job.failed_runs+' failed runs';
    const runs=el('data-job-runs'),fails=el('data-job-fails'),fileCount=el('data-job-files'),byteCount=el('data-job-bytes');
    if(runs)runs.textContent=Number(job.successful_runs||0).toLocaleString();
    if(fails)fails.textContent=Number(job.failed_runs||0).toLocaleString();
    if(fileCount)fileCount.textContent=files.toLocaleString();
    if(byteCount)byteCount.textContent=formatBytes(bytes);
    const progress=el('data-activity-running');
    if(progress){progress.textContent=job.running?'Running':job.error?'Last run failed':'Idle';progress.classList.toggle('off',!!job.error);}
    const liveProgress=job.progress||{},progressBox=el('data-job-progress');
    if(progressBox){
      const stage=el('data-job-phase'),bar=el('data-job-bar'),info=el('data-job-progress-detail');
      const running=!!liveProgress.running,done=Number(liveProgress.processed||0),total=Number(liveProgress.total||0);
      // A sync card should be quiet when it isn't doing work.
      // Persistent errors have their own separate, visible message.
      progressBox.classList.toggle('hidden',!running);
      if(stage)stage.textContent=running?'Syncing · '+(liveProgress.phase||'working'):liveProgress.phase==='failed'?'Last run failed':liveProgress.phase==='complete'?'Last run complete':'Idle';
      if(bar){if(running&&total===0){bar.removeAttribute('value');}else{bar.value=total>0?Math.min(100,Math.round(done*100/total)):(liveProgress.phase==='complete'?100:0);}}
      if(info)info.textContent=(total>0?done+' / '+total+' entries processed · ':'')+
        Number(liveProgress.uploads||0)+' uploaded, '+Number(liveProgress.downloads||0)+' downloaded · '+
        formatBytes(Number(liveProgress.upload_bytes||0)+Number(liveProgress.download_bytes||0))+
        (liveProgress.current?' · '+liveProgress.current:'')+(liveProgress.error?' · '+liveProgress.error:'');
    }
    const detail=el('data-job-detail');
    if(detail){
      const duration=Number(job.last_duration_ms||0)/1000;
      const transferred=Number(job.last_upload_bytes||0)+Number(job.last_download_bytes||0);
      const rate=duration>0?(transferred/duration):0;
      detail.textContent='Latest run: '+(job.last_run||'Never')+' · '+(job.last_uploads||0)+' uploads / '+(job.last_downloads||0)+' downloads · '+formatBytes(transferred)+' · '+duration.toFixed(1)+'s · '+formatBytes(rate)+'/s average'+(job.error?' · '+job.error:'');
    }
  }
  for(const share of data.shares||[]){
   for(const el of document.querySelectorAll('[data-share-mounted]')){
    if(el.dataset.shareMounted!==share.id)continue;
    el.textContent=share.direct?'NAS connection saved':share.mounted?'Mounted':'Not mounted';
    el.classList.toggle('off',!share.direct&&!share.mounted);
   }
  }
 }catch(_e){/* Keep the current UI intact through temporary connection loss. */}
 finally{dashboardPollBusy=false;}
}
refreshDashboardStatus();
setInterval(refreshDashboardStatus,2500);
document.addEventListener('visibilitychange',()=>{if(!document.hidden)refreshDashboardStatus()});
window.addEventListener('pageshow',event=>{if(event.persisted)refreshDashboardStatus();});
document.addEventListener('keydown',e=>{if(e.key==='Escape')closePicker()});if(location.hash){let page=location.hash.slice(1);if($('view-'+page))navigate(page)}
</script></body></html>`
