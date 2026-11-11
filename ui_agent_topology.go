// Topology augments the approved Syncria workspace, without adding a second
// administration shell or inventing edges between unrelated storage devices.
package main

import "strings"

func init() {
	patches := [][2]string{
		{`      <button type="button" data-view="accounts"`, `      <button type="button" data-view="topology" class="" aria-label="Sync Topology"><span class="nav-icon">◇</span><span class="nav-copy"><strong>Sync Map</strong><small>See your connections</small></span></button>
      <button type="button" data-view="accounts"`},
		{`</style></head><body>`, `.sync-topology-list{display:flex;flex-direction:column;gap:12px}
.sync-topology-row{display:grid;grid-template-columns:minmax(0,1fr) minmax(95px,auto) minmax(0,1fr);gap:12px;align-items:center;padding:14px;border:1px solid #dbe5ed;border-radius:12px}
.sync-topology-endpoint{min-width:0;padding:12px;border:1px solid #c7d8e5;border-radius:10px;background:var(--surface-soft,#f0f6fb)}
.sync-topology-endpoint strong,.sync-topology-endpoint small{display:block;overflow-wrap:anywhere}
.sync-topology-arrow{text-align:center;font-weight:700;color:var(--accent,#406b91)}
.sync-topology-arrow small{display:block;font-weight:400;color:#66788a;overflow-wrap:anywhere}
.sync-topology-nodes{display:flex;flex-wrap:wrap;gap:10px}
.sync-topology-nodes .sync-topology-endpoint{flex:1 1 180px}
@media(max-width:640px){.sync-topology-row{grid-template-columns:minmax(0,1fr)}.sync-topology-arrow{transform:none}}
</style></head><body>`},
		{`<section class="view" id="view-devices"`, `<section class="view" id="view-topology" aria-labelledby="syncTopologyHeading">
 <div class="workspace-intro"><div><p>See how approved devices, NAS storage and cloud accounts are connected through your actual saved syncs. Device status means it recently checked in, not that a transfer is running.</p></div></div>
 <section class="panel"><div class="row agent-heading"><h2 id="syncTopologyHeading">Your sync map</h2><button id="syncTopologyRefresh" type="button" class="secondary">Refresh</button></div>
 <div id="syncTopologyEdges" class="sync-topology-list" aria-live="polite">Loading saved connections…</div>
 </section>
 <section class="panel"><h2>Connected endpoints</h2><div id="syncTopologyNodes" class="sync-topology-nodes" aria-live="polite"></div></section>
 <p id="syncTopologyWarnings" class="agent-feedback" role="status"></p>
</section>
<section class="view" id="view-devices"`},
		{`function navigate(view){`, `async function refreshSyncTopology(){
 if(currentView!=='topology')return;
 const edgesBox=$('syncTopologyEdges'),nodesBox=$('syncTopologyNodes'),warnings=$('syncTopologyWarnings');
 try{
  const response=await fetch('/api/v1/sync/topology',{credentials:'same-origin',cache:'no-store'});
  if(response.status===401){window.location.replace('/?msg=Session%20expired.');return;}
  if(!response.ok)throw new Error('Topology request failed (HTTP '+response.status+').');
  const graph=await response.json();
  const names=new Map((graph.nodes||[]).map(node=>[node.id,node]));
  edgesBox.replaceChildren();nodesBox.replaceChildren();
  if(!graph.edges||graph.edges.length===0)edgesBox.append(agentText('p','No supported sync connections are currently configured.','empty'));
  for(const edge of graph.edges||[]){
   const from=names.get(edge.source_id),to=names.get(edge.target_id);
   if(!from||!to)continue;
   const row=document.createElement('div');row.className='sync-topology-row';
   const left=document.createElement('div');left.className='sync-topology-endpoint';
   left.append(agentText('strong',from.name),agentText('small',from.kind+' · '+from.status.replaceAll('_',' ')));
   const link=document.createElement('div');link.className='sync-topology-arrow';
   link.append(agentText('span',edge.mode==='bidirectional'?'↔':'→'),agentText('small',edge.name+' · '+edge.status));
   const right=document.createElement('div');right.className='sync-topology-endpoint';
   right.append(agentText('strong',to.name),agentText('small',to.kind+' · '+to.status.replaceAll('_',' ')));
   row.append(left,link,right);edgesBox.append(row);
  }
  for(const node of graph.nodes||[]){
   const tile=document.createElement('div');tile.className='sync-topology-endpoint';
   tile.append(agentText('strong',node.name),agentText('small',node.kind+' · '+node.status.replaceAll('_',' ')));
   nodesBox.append(tile);
  }
  warnings.textContent=(graph.warnings||[]).join(' · ');
 }catch(error){edgesBox.replaceChildren(agentText('p','Unable to load sync map: '+error.message,'empty'));nodesBox.replaceChildren();warnings.textContent='';}
}
$('syncTopologyRefresh').addEventListener('click',refreshSyncTopology);

function navigate(view){`},
		{`  devices:['Connected devices','DEVICES'],`, `  devices:['Connected devices','DEVICES'],
  topology:['Sync map','SYNC MAP'],`},
		{` if(view==='devices')refreshAgentDevices();`, ` if(view==='devices')refreshAgentDevices();
 if(view==='topology')refreshSyncTopology();`},
	}
	for _, patch := range patches {
		if strings.Count(newUI, patch[0]) != 1 {
			panic("Syncria topology integration marker not unique: " + patch[0][:min(len(patch[0]), 48)])
		}
		newUI = strings.Replace(newUI, patch[0], patch[1], 1)
	}
}
