package main

// v051Theme is the shared visual system used by the existing dashboard.
// The name remains stable because prior UI tests reference it.
// All pages use the same type scale, layout grid, form language and spacing.
const v051Theme = `
/* Syncria workspace — shared design tokens */
:root{
 color-scheme:light;--bg:#f5f7fb;--surface:#fff;--surface2:#f7faff;
 --line:#e0e8f2;--ink:#183550;--sub:#647a91;--accent:#155fd2;
 --accent2:#edf4ff;--warn:#9e671b;--bad:#a82c47;
 --radius:16px;--shadow:0 4px 16px rgba(25,57,94,.045);
 font:14px/1.55 system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif
}
body{background:var(--bg);color:var(--ink);line-height:1.6}
main.app-layout{max-width:none;grid-template-columns:252px minmax(0,1fr);width:100%}
.app-layout .sidebar{
 position:sticky;top:0;height:100vh;display:flex;flex-direction:column;gap:27px;
 padding:31px 17px 22px;background:#fff;border-right:1px solid #e1e8f2;
 box-shadow:none;overflow-y:auto;z-index:8
}
.app-layout .brand{padding:0 12px}
.app-layout .brand img{width:192px!important;max-width:100%;height:auto}
.tiny{margin:0 0 10px;padding:0 14px;font-size:10px;letter-spacing:1.4px;color:#8393a7}
.app-layout .nav{display:grid;gap:5px}
.app-layout .nav button{
 display:flex;align-items:center;gap:13px;min-height:46px;width:100%;
 border:1px solid transparent;border-radius:11px;padding:10px 13px;
 background:transparent;color:#536c87;font-size:13px;font-weight:650;
 transition:background .15s,border-color .15s,color .15s,box-shadow .15s
}
.app-layout .nav button:hover:not(.active){background:#f3f7fc;color:#163e6d}
.app-layout .nav button.active{
 background:#eaf3ff;border-color:#d0e1f7;color:#155bc1;
 box-shadow:inset 3px 0 0 #1767de
}
.nav-icon{width:23px;min-width:23px;height:23px;display:inline-flex;align-items:center;justify-content:center}
.nav-icon svg{display:block}
.sidebar-bottom{margin-top:auto;padding:18px 12px 0;border-top:1px solid var(--line);font-size:12px;color:#72859b}
.sidebar-bottom p{margin:3px 0 0}
.content{width:100%;max-width:1520px;margin:0 auto;min-width:0;padding:36px clamp(22px,3.6vw,64px) 80px}
.topline{display:flex;align-items:center;justify-content:space-between;gap:18px;padding-bottom:22px;margin-bottom:28px;border-bottom:1px solid var(--line)}
.eyebrow{font-size:10px;letter-spacing:1.6px;text-transform:uppercase;font-weight:800;color:#2366bd}
.topline h1{margin:5px 0 0;font-size:clamp(25px,2.6vw,34px);letter-spacing:-1px;line-height:1.25;color:#173656}
.logout{border-radius:10px;padding:9px 15px;color:#285476;background:white;border:1px solid #d1dfee}
.lead{color:#627991;max-width:910px;line-height:1.75;margin:0 0 23px;font-size:13px}
.view.active{animation:workspaceEnter .18s ease-out}
@keyframes workspaceEnter{from{opacity:.55;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}

/* Shared cards, headings and spacing */
.stat,.panel,.quick-start,.relationship-card,.workspace-hero{
 background:#fff;border:1px solid var(--line);border-radius:var(--radius);box-shadow:var(--shadow)
}
.panel{padding:23px 25px;margin-bottom:18px;min-width:0}
.panel h2,.workspace-toolbar h2{font-size:17px;font-weight:760;letter-spacing:-.35px;color:#173856;margin:0}
.panel h3{font-size:14px;font-weight:740;color:#1d486f}
.panel-head{display:flex;justify-content:space-between;align-items:center;gap:15px;flex-wrap:wrap;margin:0 0 19px}
.panel-head h2{margin:0}
.panel p{line-height:1.65}
.section-eyebrow{
 color:#72869e;font-size:10px;text-transform:uppercase;font-weight:790;
 letter-spacing:1.2px;margin:21px 0 10px;padding-left:3px
}
.workspace-toolbar{
 display:flex;align-items:center;justify-content:space-between;gap:16px;flex-wrap:wrap;
 padding:8px 0 19px;margin:0
}
.workspace-toolbar p{font-size:12px;color:#6a8097;margin:5px 0 0}
.workspace-hero{
 display:flex;align-items:center;justify-content:space-between;gap:28px;
 padding:29px 31px;margin-bottom:20px;
 background:linear-gradient(110deg,#fff 0%,#fff 60%,#f2f7ff 100%);
}
.workspace-hero h2{font-size:21px;line-height:1.3;letter-spacing:-.55px;margin:0 0 8px;color:#193958}
.workspace-hero p{max-width:650px;font-size:13px;line-height:1.7;margin:0;color:#627a93}
.workspace-hero .primary{white-space:nowrap}
.split{display:grid;grid-template-columns:1.08fr .92fr;gap:18px;align-items:start}
.split>.panel{min-width:0}
.stats{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px;margin-bottom:19px}
.stats-setup{grid-template-columns:repeat(4,minmax(0,1fr))}
.stats-activity-overview{grid-template-columns:repeat(3,minmax(0,1fr))}
.stat{padding:19px 21px;min-height:112px}
.stat small{font-size:11px;color:#698098;line-height:1.55;display:block}
.stat strong{display:block;color:#183a5d;font-weight:760;font-size:28px;line-height:1.3;letter-spacing:-.85px;margin:5px 0;font-variant-numeric:tabular-nums;overflow-wrap:anywhere}
.stats-compact{grid-template-columns:repeat(4,minmax(0,1fr));gap:9px;margin:13px 0}
.stats-compact .stat{padding:12px 13px;min-height:auto;background:#f8fbff;box-shadow:none;border-radius:10px}
.stats-compact .stat strong{font-size:19px;letter-spacing:-.4px}
.row{display:flex;align-items:center;flex-wrap:wrap;gap:10px}
.entry{display:flex;justify-content:space-between;align-items:center;gap:18px;border-top:1px solid #e7edf5;padding:17px 0;min-width:0}
.entry:first-of-type{border-top:0}
.entry>div{min-width:0}.entry strong{font-size:13px;color:#173959;overflow-wrap:anywhere}
.entry p{color:#657d96;margin:4px 0;font-size:12px;line-height:1.65;overflow-wrap:anywhere}
.entry .row{justify-content:flex-start}
.empty{
 padding:34px 22px;background:#f9fbfe;border:1px dashed #ccdbed;border-radius:12px;
 text-align:center;font-size:13px;color:#73859a
}
.empty strong{font-size:14px;color:#224364;display:block;margin:0 0 6px}
.quiet-label{display:inline-flex;color:#73859a;font-size:11px;font-weight:670;letter-spacing:.25px}
.card-label{color:#677d95;font-size:11px;letter-spacing:.1px}
.pill,.status,.relationship-kind{
 font-size:11px;font-weight:740;line-height:1.4;display:inline-flex;align-items:center;
 padding:5px 10px;border-radius:99px;border:1px solid #c9ddf5;background:#edf5ff;color:#2162a9
}
.status{background:#ebf8f2;color:#216d4e;border-color:#c6e8d6}
.status.off{background:#fff6e9;color:#996718;border-color:#f2dbb7}
.notice{padding:14px 17px;border-radius:11px;background:#edf5ff;color:#1c568f;border:1px solid #cce0f7}
.notice.bad{background:#fff1f3;color:#a22f4b;border-color:#f0c7d0}

/* Consistent controls and keyboard accessibility */
.primary,.secondary,.danger,.logout{
 min-height:40px;display:inline-flex;align-items:center;justify-content:center;
 border-radius:10px;font-size:12px;font-weight:720;padding:10px 15px;
 text-decoration:none;line-height:1.4;transition:background .14s,border-color .14s,box-shadow .14s
}
.primary{background:#1765d6;color:#fff;border:1px solid #1765d6;box-shadow:0 3px 8px #1b67ce18}
.primary:hover{background:#0e55ba;border-color:#0e55ba;filter:none}
.secondary{background:#fff;color:#29547a;border:1px solid #ccdaea}
.secondary:hover,.logout:hover{background:#f3f8fe;border-color:#a9c5e5}
.danger{background:#fff;color:#aa3049;border:1px solid #edcbd1}
.danger:hover{background:#fff3f4}
button:focus-visible,a:focus-visible,summary:focus-visible,input:focus-visible,select:focus-visible{
 outline:3px solid #93c1fb;outline-offset:3px
}
button:disabled{opacity:.48}
.field{display:block;font-size:12px;font-weight:690;color:#2c5173;margin:0 0 17px;min-width:0}
.field-help{display:block;font-size:12px;line-height:1.6;color:#6b819a;font-weight:430;margin:4px 0 6px}
.field input,.field select{
 display:block;width:100%;min-height:43px;padding:10px 12px;border-radius:10px;
 background:#fff;color:#1b3b5c;border:1px solid #cad9ea;margin:7px 0 0
}
.field input:focus,.field select:focus{outline:3px solid #cce1fc;outline-offset:0;border-color:#367dd1}
input[type="checkbox"]{width:16px;height:16px;accent-color:#1765d6;vertical-align:middle}
.settings-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:6px 20px}
.selected,.review-grid>div{background:#f7faff;border:1px solid #dfebf8;border-radius:12px}
.selected{padding:15px 17px;gap:14px;display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;min-width:0}
.selected strong,.review-grid strong{overflow-wrap:anywhere}
.review-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:11px}
.review-grid>div{padding:14px 16px;min-width:0}
.review-grid strong{display:block;color:#18446c;margin-top:6px;font-size:13px}
.step-indicator{display:flex;gap:8px;flex-wrap:wrap;margin:0 0 27px}
.step-indicator span{border:1px solid #dce6f2;background:#f5f8fd;border-radius:10px;padding:10px 13px;font-size:12px;color:#7489a0}
.step-indicator span.active{background:#eaf4ff;color:#145cbd;border-color:#b7d7fa;font-weight:730}
.wizard-status{line-height:1.65;padding:12px 15px;border-radius:10px;background:#f2f8ff;border:1px solid #d8e6f5;margin:15px 0;color:#315a7e}
.wizard-status.good{background:#e9f7f0;border-color:#c9e8d8;color:#287252}
.wizard-status.bad{background:#fff1f3;border-color:#f2c8d0;color:#a32f48}
.plain-guide{font-size:13px;line-height:1.7;color:#2c567e;background:#f0f7ff;border:1px solid #d2e4f8;border-radius:12px;padding:13px 17px;margin:0 0 20px}
.plain-guide strong{color:#184e87}
.simple-details{border:1px solid #e0e9f4;border-radius:11px;padding:13px 16px;margin:12px 0;background:#f9fbfe;color:#325878}
.simple-details summary{font-weight:700;font-size:12px;cursor:pointer}
.simple-details dl{display:grid;grid-template-columns:minmax(80px,1fr) 3fr;gap:8px 16px;margin:15px 0 0;font-size:12px}
.simple-details dt{font-weight:750;color:#254b70}
.simple-details dd{margin:0;color:#667d96}
.simple-details .field{margin-top:12px}
.picker{background:#10294870}
.picker-box{background:#fff;border:1px solid #d1e0f1;border-radius:17px;box-shadow:0 24px 80px #182e4a36}
.picker-head,.picker-actions,.picker-footer{border-color:#e4ebf4}
.folder-row{color:#1d476e;border-radius:9px}
.folder-row:hover{background:#edf5ff}
code{color:#1162bc;overflow-wrap:anywhere}

/* Home: onboarding takes the lead only until a sync has been configured */
.quick-start{padding:23px 25px 18px;margin-bottom:20px}
.quick-start-heading{display:flex;justify-content:space-between;gap:17px;align-items:center;flex-wrap:wrap;margin-bottom:16px}
.quick-start-heading h2{margin:0;font-size:17px;color:#193c60}
.quick-start-heading p{margin:0;font-size:12px;color:#72879b}
.quick-start-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}
.quick-step{display:flex;flex-direction:column;align-items:flex-start;padding:20px;background:#f7faff;border:1px solid #e0eaf5;border-radius:12px;min-width:0}
.step-number{display:grid;place-items:center;width:30px;height:30px;border-radius:50%;background:#e4f0ff;color:#155dc7;font-size:13px;font-weight:800;margin:0 0 12px}
.quick-step h3{margin:0 0 6px;font-size:13px;font-weight:750;color:#1e436a}
.quick-step p{font-size:12px;color:#677f98;line-height:1.65;flex:1;margin:0 0 16px}
.quick-step button{margin-top:auto}
.onboarding-disclosure{margin:0 0 19px;padding:12px 18px;border:1px solid #e1eaf4;border-radius:12px;background:#fff}
.onboarding-disclosure>summary{font-size:12px;color:#376b9f;font-weight:730;cursor:pointer}
.onboarding-disclosure[open]>summary{margin-bottom:15px}
.onboarding-disclosure .quick-start{margin:0;box-shadow:none;background:transparent;border:none;padding:8px 0}
.onboarding-disclosure .quick-start-heading{display:none}
.home-panels>.panel{margin-bottom:0}

/* Sync list: each relationship is ONE card, not a card nested in a card */
.sync-library{margin:8px 0 25px}
.sync-library>.workspace-toolbar{border-bottom:1px solid #e0e8f3;margin-bottom:16px;padding-bottom:20px}
.relationship-card{display:flex;flex-direction:column;align-items:stretch;gap:0;padding:21px 24px;margin:0 0 14px;border:1px solid #dce7f3}
.relationship-card>.row{padding-top:16px;margin-top:17px;border-top:1px solid #e8eef5}
.relationship-card .relationship-kind{border-radius:8px;margin:0 8px 8px 0}
.relationship-card>div:first-child>strong{font-size:17px;font-weight:765;letter-spacing:-.25px}
.relationship-card .entry{background:transparent}
.relationship-card p{margin:6px 0;color:#68809a}
.relationship-metadata{display:flex;gap:8px;flex-wrap:wrap;margin:12px 0 9px;color:#607995}
.relationship-metadata span{background:#f5f8fc;border:1px solid #e4ebf4;padding:5px 9px;border-radius:8px;font-size:11px}
.relationship-actions .danger{margin-left:auto}
.relationship-alert{border-radius:12px;padding:17px;background:#fff8ec;border:1px solid #f0dcb9;margin:12px 0}
.relationship-alert h3{margin:0 0 6px;font-size:13px}
.sync-live-progress{margin:12px 0 2px;padding:16px;border-radius:11px;border:1px solid #d8e6f5;background:#f8fbff}

/* Activity: readable groups without extra nested white blocks */
#view-activity>.stats{grid-template-columns:repeat(3,minmax(0,1fr))}
#view-activity .panel>.entry{align-items:stretch}
#view-activity .panel>.entry>div{width:100%}
#view-storage>.row{justify-content:flex-start;margin-bottom:19px}
#view-storage .panel,#view-accounts .panel,#view-settings .panel{margin-top:0}

/* Update idle is a clean screen. Details exist only while running or failed. */
.update-workspace{display:grid;grid-template-columns:minmax(0,1.55fr) minmax(290px,1fr);gap:18px;align-items:start}
.update-workspace>.panel{margin-bottom:16px}
.update-workspace>div{min-width:0}
.update-hero{display:flex;justify-content:space-between;align-items:center;gap:16px;padding:18px 19px;background:#f7faff;border:1px solid #e0ebf8;border-radius:12px;margin:0 0 18px}
.update-hero strong{font-size:24px;color:#1b4069;letter-spacing:-.7px;display:block}
.update-hero p{font-size:12px;color:#647c96;margin:4px 0}
.update-revisions{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;margin-bottom:18px}
.update-revisions>div{background:#f8fbff;border:1px solid #e3ecf6;border-radius:11px;padding:14px 15px;min-width:0}
.update-revisions small{display:block;font-size:10px;color:#6a8099;font-weight:760;text-transform:uppercase;letter-spacing:.65px}
.update-revisions strong{display:block;font-size:15px;margin-top:5px;color:#1d456f;overflow-wrap:anywhere}
.update-progress-panel{background:#f7faff;border:1px solid #d7e5f5;border-radius:12px;margin:0;padding:19px}
.update-progress-panel .update-progress-meta{display:flex;align-items:center;justify-content:space-between;gap:10px;flex-wrap:wrap}
.update-progress-panel .update-progress-meta strong{font-size:14px;color:#193f66}
.update-progress-panel progress{width:100%;accent-color:#1765d5;height:18px;margin:16px 0 8px}
.update-connection{border-radius:10px;background:#ecf5ff;border:1px solid #d1e2f7;color:#4d7093;padding:11px 13px;font-size:12px}
.update-connection.lost{background:#fff9ee;color:#98621d;border-color:#efd7b0}
.update-measures{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px;margin:13px 0}
.update-measures>div{padding:12px 13px;background:#fff;border:1px solid #e1eaf5;border-radius:10px;min-width:0}
.update-measures small{display:block;font-size:11px;color:#6e839b}
.update-measures strong{display:block;color:#1c426d;font-size:12px;overflow-wrap:anywhere}
.update-steps{list-style:none;display:flex;gap:7px;flex-wrap:wrap;margin:14px 0;padding:0}
.update-steps li{background:#fff;color:#6a819a;border:1px solid #e0e9f4;border-radius:8px;padding:6px 10px;font-size:11px}
.update-steps li.active{background:#eaf3ff;color:#155db9;border-color:#b4d3f8}
.update-steps li.done{background:#eaf8f1;color:#277355;border-color:#c7e7d6}
.update-steps li.failed{background:#fff2f3;color:#a43049;border-color:#eec6d0}
.update-timeline{list-style:none;margin:0;padding:0}
.update-timeline li{border-left:2px solid #dce6f2;position:relative;margin-left:8px;padding:0 0 16px 18px;font-size:12px;color:#6a8098}
.update-timeline li:before{content:"";position:absolute;left:-5px;top:5px;width:8px;height:8px;border-radius:50%;background:#b9cfe3}
.update-timeline li.active:before{background:#1765d6;box-shadow:0 0 0 4px #ebf3ff}
.update-timeline li.failed:before{background:#a93651}
.update-timeline li strong{display:block;color:#244769;font-size:12px}
.update-timeline li small{font-size:11px;color:#798fa5}
.update-explainer{display:grid;gap:17px;margin-top:17px}
.update-explainer>div{display:flex;align-items:flex-start;gap:13px}
.update-explainer>div>span{width:29px;height:29px;min-width:29px;border-radius:9px;display:grid;place-items:center;background:#eaf3ff;color:#145cbf;font-size:12px;font-weight:800}
.update-explainer p{margin:0}
.update-explainer strong{font-size:12px;color:#244768;display:block}
.update-explainer small{display:block;margin-top:3px;font-size:11px;color:#71869d}
.advanced-update{margin:0}
.advanced-update>summary{font-size:12px;font-weight:760;color:#315d83;cursor:pointer}
.advanced-update[open]>summary{margin-bottom:17px}
.advanced-update-body{border-top:1px solid #e7eef5;padding-top:17px}
.advanced-update-body h2{margin:0 0 9px}
.update-live-banner{display:flex;align-items:center;justify-content:space-between;gap:14px;background:#eaf3ff;border:1px solid #bad8f7;border-radius:11px;padding:13px 16px;margin-bottom:18px}
.update-live-banner strong{color:#1d558b;font-size:12px}
.update-live-banner small{color:#5c7798}
.update-alert{padding:13px;border:1px solid #edc4ca;border-radius:10px;background:#fff3f5;color:#ac324c;font-size:12px}

/* Compact screens, flexible columns and reachable actions */
@media(max-width:1190px){
 .update-workspace{grid-template-columns:1fr}
 .stats{grid-template-columns:repeat(2,minmax(0,1fr))}
 .stats-setup{grid-template-columns:repeat(4,minmax(0,1fr))}
 .quick-start-grid{grid-template-columns:repeat(2,minmax(0,1fr))}
}
@media(max-width:980px){
 main.app-layout{grid-template-columns:minmax(0,1fr)}
 .app-layout .sidebar{position:static;height:auto;overflow:visible;display:flex;flex-direction:column;padding:16px 18px;gap:13px}
 .app-layout .brand img{width:159px!important}
 .app-layout .nav{display:flex;gap:6px;flex-wrap:wrap}
 .app-layout .nav button{min-height:39px;width:auto;padding:9px 11px;gap:8px}
 .nav-icon{width:18px;min-width:18px}
 .nav-icon svg{width:17px;height:17px}
 .sidebar-bottom,.tiny{display:none}
 .content{padding:25px 19px 60px}
 .split{grid-template-columns:1fr}
 .home-panels>.panel{margin-bottom:17px}
}
@media(max-width:680px){
 .content{padding:19px 14px 55px}
 .topline h1{font-size:26px}
 .panel{padding:18px}
 .workspace-hero{padding:22px;flex-direction:column;align-items:stretch;gap:16px}
 .workspace-hero h2{font-size:19px}
 .workspace-hero .primary{align-self:flex-start}
 .stats,.stats-setup,.stats-activity-overview,#view-activity>.stats{grid-template-columns:repeat(2,minmax(0,1fr));gap:10px}
 .stat{padding:14px;min-height:98px}
 .stat strong{font-size:24px}
 .quick-start{padding:18px}
 .quick-start-grid,.settings-grid,.review-grid,.update-revisions{grid-template-columns:1fr}
 .entry{flex-direction:column;align-items:stretch;gap:11px}
 .entry .row{width:100%}
 .relationship-card{padding:17px}
 .relationship-actions .danger{margin-left:0}
 .relationship-actions button{min-height:41px}
 .update-hero{align-items:stretch;flex-direction:column}
 .update-measures{grid-template-columns:1fr 1fr}
 .step-indicator span{padding:8px 10px}
 .simple-details dl{grid-template-columns:1fr;gap:3px}
 .simple-details dd{margin-bottom:8px}
}
@media(prefers-reduced-motion:reduce){.view.active{animation:none}button{transition:none!important}}
`
