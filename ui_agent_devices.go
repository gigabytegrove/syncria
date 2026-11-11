package main

import "strings"

// Device management is an additive template extension. Exact-match guards
// intentionally fail closed when the established UI structure changes.
// This preserves the existing dashboard and requires no second shell page.
func init() {
	patches := [][2]string{
		{`      <button type="button" data-view="accounts"`, `      <button type="button" data-view="devices" class="" aria-label="Connected Devices"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="4" width="18" height="12" rx="2"/><path d="M8 20h8M12 16v4"/></svg></span><span class="nav-copy"><strong>Devices</strong><small>Computers & agents</small></span></button>
      <button type="button" data-view="accounts"`},
		{`</style></head><body>`, `.agent-invite-form{display:flex;align-items:end;gap:12px;flex-wrap:wrap}
.agent-invite-form .field{flex:1;min-width:220px}
.agent-invitation-details{background:var(--surface-soft,#f0f6fb);border:1px solid #cbddea;border-radius:12px;padding:17px;margin-top:16px}
.agent-invitation-details input{display:block;width:100%;margin:10px 0 12px;padding:12px;font-family:ui-monospace,SFMono-Regular,Consolas,monospace;letter-spacing:.025em;overflow-wrap:anywhere}
.agent-feedback{min-height:22px;margin:10px 0 0}
.agent-feedback.bad{color:#b42318}.agent-feedback.good{color:#236b44}
.agent-heading{justify-content:space-between}
.agent-list{display:flex;flex-direction:column;gap:0}
.agent-item{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:14px 0;border-top:1px solid #d8e2ec}
.agent-item-main{display:flex;flex-direction:column;gap:4px;min-width:0}
.agent-item small{display:block;color:#617891;overflow-wrap:anywhere}
.agent-item strong{overflow-wrap:anywhere}
.agent-activity{justify-content:flex-start}
@media(max-width:640px){.agent-item{align-items:flex-start;flex-direction:column}.agent-invite-form button{width:100%}}

</style></head><body>`},
		{`<section class="view" id="view-updates">`, `<section class="view" id="view-devices" aria-labelledby="agentPageHeading">
 <div class="workspace-intro"><div><p>Connect trusted computers to Syncria. Connected devices can report their status; folder synchronization will be available when the agent transfer system is ready.</p></div></div>
 <section class="panel">
  <h2 id="agentPageHeading">Connect a device</h2>
  <p>Enter a name for your computer and create a temporary pairing code. The code does not give Syncria access to your files.</p>
  <form id="agentInvitationForm" class="agent-invite-form">
   <label class="field">Device name<input id="agentInvitationName" type="text" maxlength="128" required placeholder="Living room PC" autocomplete="off"></label>
   <button class="primary" id="agentInvitationSubmit" type="submit">Create pairing code</button>
  </form>
  <div id="agentInvitationDetails" class="agent-invitation-details hidden" aria-live="polite">
   <strong>Pairing code</strong><p>Enter this code in the Syncria agent on your computer. It expires in 10 minutes and is shown only once.</p>
   <input id="agentInvitationToken" type="text" readonly aria-label="One-time pairing code" autocomplete="off" spellcheck="false">
   <div class="row"><button type="button" class="secondary" id="agentCopyToken">Copy code</button><button type="button" class="primary" id="agentApproveCurrent">Approve this device</button><button type="button" class="secondary" id="agentHideToken">Hide code</button></div>
  </div>
  <p id="agentInvitationFeedback" class="agent-feedback" role="status" aria-live="polite"></p>
 </section>
 <section class="panel"><div class="row agent-heading"><h2>Connected devices</h2><button type="button" id="agentRefresh" class="secondary">Refresh</button></div><p>Only approved, enrolled devices appear here. Recent check-ins indicate that an agent has contacted Syncria; they do not confirm an active synchronization job.</p><div id="agentDeviceList" class="agent-list" aria-live="polite">Loading connected devices…</div></section>
 <section class="panel"><h2>Pairing requests</h2><p>Review any outstanding invitations before allowing an agent to enroll.</p><div id="agentInvitationList" class="agent-list" aria-live="polite">Loading requests…</div></section>
 <section class="panel"><h2>Device activity</h2><p>Enrollment, approval, and revocation events from the Syncria controller.</p><div id="agentAuditList" class="agent-list" aria-live="polite">Loading device activity…</div></section>
</section>

<section class="view" id="view-updates">`},
		{`function navigate(view){`, `let currentAgentInvitation=null;
function agentStatus(message,bad){
 const el=$('agentInvitationFeedback');
 if(!el)return;
 el.textContent=message;
 el.classList.toggle('bad',Boolean(bad));
 el.classList.toggle('good',!bad);
}
function clearAgentPairing(){
 currentAgentInvitation=null;
 if($('agentInvitationToken'))$('agentInvitationToken').value='';
 if($('agentInvitationDetails'))$('agentInvitationDetails').classList.add('hidden');
}
async function agentRequest(url,method,data){
 const options={method:method||'GET',credentials:'same-origin',cache:'no-store',headers:{'Accept':'application/json'}};
 if(data!==undefined){options.headers['Content-Type']='application/json';options.body=JSON.stringify(data);}
 const response=await fetch(url,options);
 if(response.status===401){
  window.location.replace('/?msg=Session%20expired.%20Please%20sign%20in%20again.');
  throw new Error('Administrator session expired');
 }
 if(!response.ok){throw new Error('Request failed (HTTP '+response.status+').');}
 return response.json();
}
function agentText(tag,value,klass){
 const el=document.createElement(tag);
 el.textContent=String(value);
 if(klass)el.className=klass;
 return el;
}
function agentDate(value){
 if(!value)return 'Never';
 const date=new Date(value);
 return Number.isFinite(date.valueOf())?date.toLocaleString():'Unknown';
}
function agentButton(label,action,variant){
 const button=agentText('button',label,variant||'secondary');
 button.type='button';
 button.addEventListener('click',action);
 return button;
}
async function refreshAgentDevices(){
 if(currentView!=='devices')return;
 try{
  const [devices,invites,history]=await Promise.all([
   agentRequest('/api/v1/agents/devices'),
   agentRequest('/api/v1/agents/invitations'),
   agentRequest('/api/v1/agents/audit')
  ]);
  const deviceBox=$('agentDeviceList');deviceBox.replaceChildren();
  if(!devices.devices || !devices.devices.length)deviceBox.append(agentText('p','No devices enrolled yet. Create and approve a pairing request to connect one.','empty'));
  for(const device of devices.devices||[]){
   const row=document.createElement('div');row.className='agent-item';
   const details=document.createElement('div');details.className='agent-item-main';
   details.append(agentText('strong',device.name),agentText('small',device.platform+' · '+device.arch+' · '+device.version));
   const last=device.last_seen ? Date.now()-new Date(device.last_seen).getTime():Infinity;
   const state=device.revoked_at?'Revoked':last>=0&&last<=120000?'Recently connected':'Not recently connected';
   details.append(agentText('small',state+' · Last contact: '+agentDate(device.last_seen)));
   row.append(details);
   if(!device.revoked_at){
    row.append(agentButton('Revoke',async()=>{
     if(!window.confirm('Revoke '+device.name+'? Its existing credential will stop working.'))return;
     try{await agentRequest('/api/v1/agents/devices/'+encodeURIComponent(device.id)+'/revoke','POST');agentStatus('Device revoked.');await refreshAgentDevices();}
     catch(e){agentStatus(e.message,true);}
    }));
   }
   deviceBox.append(row);
  }
  const invitesBox=$('agentInvitationList');invitesBox.replaceChildren();
  const active=(invites.invitations||[]).filter(invite=>!invite.redeemed&&new Date(invite.expires_at).getTime()>Date.now());
  if(!active.length)invitesBox.append(agentText('p','No active pairing requests.','empty'));
  for(const invite of active){
   const row=document.createElement('div');row.className='agent-item';
   const details=document.createElement('div');details.className='agent-item-main';
   details.append(agentText('strong',invite.name),agentText('small',(invite.approved?'Approved, waiting for device':'Awaiting administrator approval')+' · Expires '+agentDate(invite.expires_at)));
   row.append(details);
   if(!invite.approved){
    row.append(agentButton('Approve',async()=>{
     try{await agentRequest('/api/v1/agents/invitations/'+encodeURIComponent(invite.id)+'/approve','POST');agentStatus('Pairing request approved.');await refreshAgentDevices();}
     catch(e){agentStatus(e.message,true);}
    },'primary'));
   }
   invitesBox.append(row);
  }
  const auditBox=$('agentAuditList');auditBox.replaceChildren();
  const events=(history.events||[]).slice(-25).reverse();
  if(!events.length)auditBox.append(agentText('p','No device activity recorded yet.','empty'));
  for(const entry of events){
   const row=document.createElement('div');row.className='agent-item agent-activity';
   row.append(agentText('strong',String(entry.kind||'').replaceAll('_',' ')),agentText('small',agentDate(entry.at)+' · '+entry.subject));
   auditBox.append(row);
  }
 }catch(error){
  for(const id of ['agentDeviceList','agentInvitationList','agentAuditList']){
   const box=$(id);if(box)box.replaceChildren(agentText('p','Unable to load device information: '+error.message,'empty'));
  }
 }
}
$('agentInvitationForm').addEventListener('submit',async event=>{
 event.preventDefault();
 const button=$('agentInvitationSubmit');button.disabled=true;
 try{
  clearAgentPairing();
  const payload=await agentRequest('/api/v1/agents/invitations','POST',{name:$('agentInvitationName').value.trim(),ttl_seconds:600});
  currentAgentInvitation=payload.invitation;
  $('agentInvitationToken').value=payload.token;
  $('agentInvitationDetails').classList.remove('hidden');
  agentStatus('Pairing code created. Approve the device after verifying your request.');
  await refreshAgentDevices();
 }catch(error){agentStatus(error.message,true);}
 finally{button.disabled=false;}
});
$('agentApproveCurrent').addEventListener('click',async()=>{
 if(!currentAgentInvitation)return;
 try{
  await agentRequest('/api/v1/agents/invitations/'+encodeURIComponent(currentAgentInvitation.id)+'/approve','POST');
  agentStatus('Device approved. The agent can now enroll using the pairing code.');
  await refreshAgentDevices();
 }catch(error){agentStatus(error.message,true);}
});
$('agentCopyToken').addEventListener('click',async()=>{
 const input=$('agentInvitationToken');
 try{
  if(!navigator.clipboard || !navigator.clipboard.writeText)throw new Error('Clipboard unavailable');
  await navigator.clipboard.writeText(input.value);
  agentStatus('Pairing code copied.');
 }catch(error){
  input.focus();input.select();
  agentStatus('Select and copy the highlighted pairing code.');
 }
});
$('agentHideToken').addEventListener('click',()=>{clearAgentPairing();agentStatus('Pairing code hidden. Create a new invitation if it was not saved.');});
$('agentRefresh').addEventListener('click',refreshAgentDevices);

function navigate(view){`},
		{`  updates:['App updates','UPDATES'],`, `  updates:['App updates','UPDATES'],
  devices:['Connected devices','DEVICES'],`},
		{` currentView=view;
 document.querySelectorAll('.view')`, ` if(currentView==='devices'&&view!=='devices')clearAgentPairing();
 currentView=view;
 document.querySelectorAll('.view')`},
		{` if(view==='updates')checkUpdate();`, ` if(view==='updates')checkUpdate();
 if(view==='devices')refreshAgentDevices();`},
	}
	for _, patch := range patches {
		if strings.Count(newUI, patch[0]) != 1 {
			panic("Syncria agent dashboard integration marker not unique: " + patch[0][:min(len(patch[0]), 48)])
		}
		newUI = strings.Replace(newUI, patch[0], patch[1], 1)
	}
}
