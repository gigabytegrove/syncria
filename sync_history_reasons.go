package main

// These are explanations of *completed* operations. The actual preview uses
// more detailed timestamp-specific reasons before any bytes are transferred.
func syncOperationReason(j *Job,action string)string{
 switch action {
 case "copy":
  return "Copied a file missing on the destination after comparing saved file metadata."
 case "replace":
  return "Existing destination file was replaced because its source or destination metadata changed since the previous successful sync."
 case "delete":
  if enabledByDefault(j.RecoveryEnabled){
   return "A previously synchronized counterpart was deleted; deletion sharing is enabled, and a recoverable copy was staged."
  }
  return "A previously synchronized counterpart was deleted and deletion sharing is enabled."
 case "one-way source wins":
  return "One-way relationship restored the authoritative source version to the destination; NAS edits are not uploaded backward."
 case "resolve first-run conflict":
  return "Different existing file versions were reconciled using the configured initial-sync preference."
 case "resolve conflict":
  return "Both versions changed; the configured conflict rule selected the destination copy."
 case "transfer started":
  return "A file transfer began; this entry alone does not mean that the transfer completed."
 case "transfer failed":
  return "The file operation failed before it could be completed. Review the run error for details."
 case "unsupported Google document — kept on Drive":
  return "Google-native content is not supported by this connector; it remains on Drive and is not copied to NAS."
 default:
  return "Recorded synchronization operation."
 }
}
