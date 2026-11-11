package main

import "strings"

// Additive integration preserves Syncria's established responsive interface.
// The source device must already be paired and its peer service running.
func init() {
	changes := [][2]string{
		{`<section class="panel"><h2>Pairing requests</h2>`, `<section class="panel" id="agentSyncSetup"><h2>Sync device files to a NAS — alpha</h2>
  <p>Create an approved, one-way device → NAS relationship. Source deletions do not delete NAS files. Different existing destination files will not be overwritten, and new relationships start paused.</p>
  <form id="agentSyncForm" class="agent-invite-form">
   <label class="field">Sync name<input id="agentSyncName" required maxlength="128" placeholder="Laptop photos to NAS" autocomplete="off"></label>
   <label class="field">Approved device<select id="agentSyncDevice" required><option value="">Select a connected device</option></select></label>
   <label class="field">Device HTTPS address<input id="agentSyncPeerURL" required placeholder="https://device.example:9774" autocomplete="url"></label>
   <label class="field">Approved source folder ID<input id="agentSyncGrant" required maxlength="32" placeholder="From agent grants list" autocomplete="off"></label>
   <label class="field">Source peer access key<input id="agentSyncPeerKey" type="password" required autocomplete="off" placeholder="Enter the device's private peer key"></label>
   <label class="field">Destination NAS<select id="agentSyncNAS" required><option value="">Choose an existing NAS</option>{{range .Shares}}{{if and (eq .Protocol "smb") (eq .Local "")}}<option value="{{.ID}}">{{.Label}}</option>{{end}}{{end}}</select></label>
   <label class="field">NAS subfolder<input id="agentSyncFolder" value="." required autocomplete="off"></label>
   <label class="field">Check every (minutes)<input id="agentSyncInterval" type="number" value="15" min="1" max="1440" required></label>
   <button class="primary" id="agentSyncSubmit" type="submit">Create paused sync</button>
  </form>
  <p id="agentSyncFeedback" role="status" aria-live="polite"></p>
  </section>
  <section class="panel"><h2>Pairing requests</h2>`},
		{`function navigate(view){`, `async function refreshAgentSyncDevices(){
  const field=$('agentSyncDevice');if(!field)return;
  try{
   const payload=await agentRequest('/api/v1/agents/devices');
   const selected=field.value;
   field.replaceChildren(new Option('Select a connected device',''));
   for(const item of payload.devices||[]){if(!item.revoked_at)field.add(new Option(item.name+' ('+item.platform+')',item.id));}
   if([...field.options].some(o=>o.value===selected))field.value=selected;
  }catch(err){$('agentSyncFeedback').textContent='Could not load paired devices: '+err.message;}
 }
 $('agentSyncForm').addEventListener('submit',async event=>{
  event.preventDefault();const button=$('agentSyncSubmit');button.disabled=true;
  try{
   const body={name:$('agentSyncName').value.trim(),device_id:$('agentSyncDevice').value,
    peer_url:$('agentSyncPeerURL').value.trim(),grant_id:$('agentSyncGrant').value.trim(),
    peer_token:$('agentSyncPeerKey').value, nas_share_id:$('agentSyncNAS').value,
    nas_folder:$('agentSyncFolder').value.trim(),interval_minutes:Number($('agentSyncInterval').value)};
   const result=await agentRequest('/api/v1/agent-sync/jobs','POST',body);
   $('agentSyncPeerKey').value='';
   $('agentSyncFeedback').textContent='Paused device sync created. Review it under My Syncs and enable it when ready.';
   window.location.hash='sync';window.location.reload();
  }catch(err){$('agentSyncFeedback').textContent='Unable to create sync: '+err.message;}
  finally{button.disabled=false;}
 });
function navigate(view){`},
		{`if(view==='devices')refreshAgentDevices();`, `if(view==='devices'){refreshAgentDevices();refreshAgentSyncDevices();}`},
		{`{{else}}FOLDER / FOLDER{{end}}`, `{{else if eq .Provider "agent_bridge"}}DEVICE / NAS{{else}}FOLDER / FOLDER{{end}}`},
		{`{{else}}Existing filesystem sync{{end}}`, `{{else if eq .Provider "agent_bridge"}}One way: approved device folder → NAS · no deletion propagation{{else}}Existing filesystem sync{{end}}`},
		{`folder synchronization will be available when the agent transfer system is ready.`, `approved folders can be sources for one-way NAS synchronization. File deletion never propagates in this alpha.`},
	}
	for _, change := range changes {
		if strings.Count(newUI, change[0]) != 1 {
			panic("Syncria agent sync UI integration marker missing or duplicated")
		}
		newUI = strings.Replace(newUI, change[0], change[1], 1)
	}
}
