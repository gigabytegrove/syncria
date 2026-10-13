package main

// v051Theme applies consistently to every existing Syncria view.
// The approved embedded logo and all established routes remain unchanged.
const v051Theme = `
/* Unified v0.5.1 dashboard: white / navy / brand blue */
:root{--bg:#f5f8fc;--surface:#fff;--surface2:#f3f7fc;--line:#dde7f2;--ink:#183451;--sub:#607894;--accent:#155fd4;--accent2:#eaf3ff;--bad:#ae3048}
body{background:#f5f8fc;color:var(--ink);line-height:1.55}
main.app-layout{max-width:none;grid-template-columns:248px minmax(0,1fr)}
.app-layout .sidebar{background:#fff;border-right:1px solid var(--line);box-shadow:none;padding:28px 16px;gap:30px;z-index:9}
.app-layout .brand{padding:0 10px}.app-layout .brand img{width:188px!important}
.app-layout .nav{gap:5px}.app-layout .nav button{gap:12px;min-height:44px;padding:10px 13px;font-size:13px;font-weight:660;color:#566f8c;border-radius:10px}
.app-layout .nav button.active{background:#eaf3ff;color:#135bc2;border-color:#c9def7;box-shadow:inset 3px 0 0 #1c66d4}
.app-layout .nav button:hover:not(.active){background:#f3f7fc}
.nav-icon{display:inline-flex;min-width:24px;height:22px;align-items:center;justify-content:center;font-size:18px;font-weight:500}
.sidebar-bottom{border-top:1px solid var(--line);padding:19px 10px 0;color:#70839a}.sidebar-bottom p{margin:4px 0 0}
.content{max-width:1590px;width:100%;margin:0 auto;padding:32px clamp(21px,3.3vw,58px) 78px}
.topline{border-bottom:1px solid #e2e9f2;padding-bottom:22px;margin-bottom:25px}.topline h1{font-size:clamp(25px,2.8vw,32px);color:#173553;letter-spacing:-.7px}
.lead{margin:0 0 23px;line-height:1.68;font-size:13px;max-width:910px}
.stats{grid-template-columns:repeat(4,minmax(0,1fr));gap:14px}.stat{min-height:116px;padding:20px}.stat strong{font-size:29px;letter-spacing:-.6px}.stat small{font-size:12px}
.stat,.panel{background:#fff;border:1px solid #dfe8f2;box-shadow:0 4px 18px #17487808;border-radius:15px}
.panel{padding:23px 24px;margin-bottom:19px}.panel h2{font-size:17px;color:#183957}.panel h3{font-size:15px;color:#244c72}.panel-head{margin-bottom:18px}
.split{grid-template-columns:1fr 1fr;gap:17px}.entry{padding:16px 0;border-color:#e5ecf5}.entry p{line-height:1.55;color:#617992}
.primary,.secondary,.danger{padding:9px 15px;min-height:39px;border-radius:10px;font-size:12px;font-weight:680;transition:background .15s,border-color .15s}
.primary{background:#1764d2;border-color:#1764d2;color:white;box-shadow:0 2px 6px #1764d120}.primary:hover{filter:none;background:#0b56c2}
.secondary{background:white;color:#2c547a;border-color:#cbdaeb}.secondary:hover{background:#f0f6ff;border-color:#bdd0e7}
.danger{background:white;border-color:#efcad0;color:#af314b}.danger:hover{background:#fff1f3}
.field{color:#385b79}.field input,.field select{background:#fff;color:#173655;border:1px solid #cedded;padding:11px 12px;border-radius:10px}
input[type=checkbox]{accent-color:#1764d2;min-width:15px;min-height:15px}.selected,.review-grid>div{background:#f7faff;border-color:#dde8f4;border-radius:12px}
.status{background:#eaf8f1;color:#257051;border-color:#bae4d0;font-weight:670}.status.off{background:#fff5e9;color:#a7751d;border-color:#efdcba}
.step-indicator span.active{background:#e8f2ff;border-color:#b7d3f4;color:#155fc4}
.notice{background:#edf5ff;color:#23578a;border-color:#c7dffe}.empty{background:#fbfdff;border-color:#d5e4f4}
.view.active{animation:syncView .17s ease-out}@keyframes syncView{from{opacity:.55;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}
.workspace-hero{display:flex;justify-content:space-between;align-items:center;gap:16px;padding:23px;background:#fff;border:1px solid #dfe8f2;border-radius:15px;margin-bottom:17px}
.workspace-hero h2{margin:0 0 6px;font-size:18px;color:#173755}.workspace-hero p{font-size:12px;color:#607890;margin:0;max-width:570px}
.relationship-card{border:1px solid #dde7f3;background:#fff;border-radius:13px;margin-bottom:14px;padding:17px 20px}
.relationship-card>.row{border-top:1px solid #eaf0f8;padding-top:14px;margin-top:13px}.relationship-card .entry{border:0;padding:0}
.relationship-kind{display:inline-flex;font-size:11px;font-weight:720;border:1px solid #d2e2f5;padding:4px 9px;border-radius:8px;color:#2461a2;background:#edf5ff;margin-bottom:7px}
.relationship-metadata{display:flex;flex-wrap:wrap;gap:8px;margin-top:9px;font-size:12px;color:#627b95}
.relationship-metadata span{background:#f6f9fc;padding:4px 8px;border:1px solid #e1eaf4;border-radius:7px}
.update-workspace{display:grid;grid-template-columns:minmax(0,1.6fr) minmax(300px,1fr);gap:18px;align-items:start}
.update-workspace>.panel{margin-bottom:18px}
.update-hero{display:flex;align-items:center;justify-content:space-between;gap:18px;padding:18px 19px;background:#f6faff;border:1px solid #dfeaf8;border-radius:13px;margin-bottom:16px}
.update-hero strong{display:block;font-size:24px;color:#173e6a;letter-spacing:-.5px}.update-hero p{font-size:12px;margin:3px 0;color:#607d99}
.update-revisions{display:grid;grid-template-columns:1fr 1fr;gap:12px;margin:14px 0}
.update-revisions>div{padding:12px 13px;border:1px solid #e2eaf5;background:#f8fbff;border-radius:10px}
.update-revisions small{display:block;font-size:11px;color:#71869c;letter-spacing:.5px;text-transform:uppercase;font-weight:700}
.update-revisions strong{font-size:15px;color:#16416f;overflow-wrap:anywhere}
.update-progress-panel{background:#f7faff;border:1px solid #d9e7f8;border-radius:13px;padding:18px}
.update-progress-panel progress{height:19px;margin:15px 0 7px;accent-color:#1665d5}
.update-progress-panel .update-progress-meta strong{font-size:16px;color:#1a4168}
.update-measures{display:flex;gap:10px;flex-wrap:wrap;margin:14px 0}.update-measures>div{min-width:155px;flex:1;background:#fff;border:1px solid #e0eaf5;border-radius:9px;padding:12px}
.update-measures small{font-size:11px;color:#71869c;display:block}.update-measures strong{display:block;color:#174475;font-size:12px;overflow-wrap:anywhere}
.update-timeline{list-style:none;display:grid;gap:0;padding:3px 0 0;margin:0}.update-timeline li{border-left:2px solid #d7e4f1;position:relative;padding:0 0 16px 17px;margin:0 0 0 8px;font-size:12px;color:#627b94}
.update-timeline li:before{content:"";width:9px;height:9px;background:#bbcde0;border-radius:50%;position:absolute;top:4px;left:-5.5px}.update-timeline li.active:before{background:#1764d2;box-shadow:0 0 0 4px #e6f0ff}.update-timeline li.failed:before{background:#b5324b}
.update-timeline li strong{color:#204469;display:block;font-weight:690;font-size:12px}.update-timeline li small{font-size:11px;color:#778ba1}
.update-connection{background:#f7faff;color:#537493;border:1px solid #d8e6f5;border-radius:9px;padding:11px 13px;font-size:12px}
.update-connection.lost{background:#fff8ec;color:#a06c1f;border-color:#efdcbc}
.update-live-banner{display:flex;justify-content:space-between;gap:12px;align-items:center;background:#eaf4ff;border:1px solid #bbd9f7;border-radius:11px;padding:12px 15px;margin-bottom:18px}.update-live-banner strong{font-size:12px;color:#175086}.update-live-banner small{color:#527495}
.update-progress-panel .update-steps li.failed{color:#a52b48}
@media(max-width:1160px){.stats{grid-template-columns:repeat(3,minmax(0,1fr))}.update-workspace{grid-template-columns:1fr}}
@media(max-width:980px){main.app-layout{grid-template-columns:1fr}.app-layout .sidebar{position:static;height:auto;padding:14px 16px;gap:14px}.app-layout .nav{display:flex;flex-wrap:wrap}.app-layout .nav button{width:auto;min-height:38px}.sidebar-bottom{display:none}.content{padding:23px 18px 55px}}
@media(max-width:680px){.stats{grid-template-columns:repeat(2,minmax(0,1fr))}.split,.settings-grid,.update-revisions{grid-template-columns:1fr}.workspace-hero,.update-hero{flex-direction:column;align-items:stretch}.panel{padding:17px}.update-workspace{gap:10px}.update-measures>div{min-width:125px}.topline h1{font-size:24px}}
`
