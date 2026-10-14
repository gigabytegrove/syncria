package main

// v051Theme is the retained Go symbol referenced by existing UI tests.
// The actual interface is Syncria's wholly new v0.6 product design system.
// It requires no external font, CSS service, or front-end build dependency.
const v051Theme = `
:root{
color-scheme:light;
--bg:#f3f6fb;--surface:#fff;--surface2:#f3f8ff;
--line:#e0e8f2;--ink:#122a49;--sub:#627994;
--accent:#1875f6;--accent2:#eaf2ff;--aqua:#10d7d4;
--nav:#0b1b36;--nav-2:#112b52;--warn:#bd7e1d;--bad:#ad2d4a;
--radius:15px;--shadow:0 4px 24px #0a2b580b;
font:14px/1.55 "Segoe UI Variable","Segoe UI",system-ui,-apple-system,sans-serif
}
*{box-sizing:border-box}
body{background:var(--bg);color:var(--ink);margin:0}
.app-layout{display:grid!important;grid-template-columns:274px minmax(0,1fr);max-width:none!important;min-height:100vh;width:100%}
.sidebar{
background:var(--nav)!important;border:0!important;color:#fff;position:sticky;top:0;
height:100vh;min-height:0;padding:25px 17px 22px;display:flex;flex-direction:column;gap:22px;
overflow-y:auto;z-index:10;box-shadow:8px 0 32px #08162c14
}
.sidebar-brand{display:grid;gap:15px;padding:0 8px}
.sidebar-logo{height:80px;border-radius:16px;background:#fff;display:flex;align-items:center;justify-content:center;padding:9px 14px}
.sidebar-logo img{display:block;width:190px;max-width:100%;height:66px;object-fit:contain}
.sidebar-edition{display:flex;align-items:center;gap:8px;color:#86a5cf;letter-spacing:1.8px;font-size:9px;font-weight:820;margin:0 7px}
.pulse-dot{display:inline-block;width:7px;height:7px;border-radius:50%;background:var(--aqua);box-shadow:0 0 0 4px #14d2d124;flex-shrink:0}
.sidebar-edition-line{height:1px;background:#375273;flex:1;margin-left:5px}
.sidebar-navigation{display:grid;gap:24px;margin-top:6px}
.nav-group{display:grid;gap:8px}
.nav-section-label{font-size:10px;letter-spacing:1.8px;color:#6182b2;font-weight:850;margin:0 12px}
.app-layout .nav{display:grid;gap:4px}
.app-layout .nav button{
width:100%;display:flex;align-items:center;text-align:left;gap:13px;
min-height:56px;padding:9px 13px;background:transparent;border:1px solid transparent;
border-radius:11px;color:#b8cbe6;box-shadow:none;transition:background .15s,color .15s,border-color .15s
}
.app-layout .nav button:hover{background:#142f56;color:#fff}
.app-layout .nav button.active{
background:#173f78;border:1px solid #326db6;color:#fff;
box-shadow:inset 3px 0 0 var(--aqua),0 5px 16px #0715293d
}
.nav-icon{height:37px;width:37px;min-width:37px;border-radius:9px;background:#203a5e;display:grid;place-items:center;color:#88add8}
.app-layout .nav button.active .nav-icon{background:#1e599f;color:#b7f8ff}
.nav-icon svg{display:block}
.nav-copy{display:grid;min-width:0;gap:1px}
.nav-copy strong{font-size:13px;font-weight:720;line-height:1.25}
.nav-copy small{font-size:10px;font-weight:500;color:#7f9dc2;line-height:1.5}
.app-layout .nav button.active .nav-copy small{color:#a8d2fa}
.sidebar-bottom{margin-top:auto;padding:16px 8px 0;border-top:1px solid #284361;color:#eaf4ff}
.sidebar-presence{display:flex;align-items:center;gap:11px}
.sidebar-presence-icon{width:35px;height:35px;background:#153f70;border-radius:10px;display:grid;place-items:center;color:var(--aqua)}
.sidebar-presence-icon svg{width:17px;height:17px}
.sidebar-presence strong,.sidebar-presence small{display:block}
.sidebar-presence strong{font-size:12px;font-weight:740}
.sidebar-presence small{font-size:10px;color:#85a6cd}
.sidebar-version{display:flex;justify-content:space-between;align-items:center;border-top:1px solid #284361;margin-top:15px;padding:13px 3px 0;color:#698db8;letter-spacing:1.5px;font-size:9px;font-weight:800}
.sidebar-version strong{font-size:11px;color:#c4dbf2;letter-spacing:0;font-weight:720}
.content{padding:31px clamp(26px,4vw,68px) 85px;min-width:0;width:100%;max-width:1680px;margin:0 auto}
.topline{
border:0!important;margin:0 0 34px;padding:0 0 24px;display:flex;align-items:center;
justify-content:space-between;gap:18px;border-bottom:1px solid #dee6f1!important
}
.page-heading{min-width:0}
.eyebrow{display:flex;gap:11px;align-items:center;font-size:10px;letter-spacing:1.55px;font-weight:800;color:#6f879e}
.breadcrumb-home{color:#1570de}
.breadcrumb-divider{color:#b4c1d1}
.topline h1{font-size:clamp(27px,2.8vw,38px);color:#152e50;line-height:1.2;letter-spacing:-1.25px;margin:8px 0 0;font-weight:830}
.header-actions{display:flex;align-items:center;gap:16px;flex-wrap:wrap}
.header-version{font-size:11px;color:#69829c;font-weight:720}
.logout{background:#fff;color:#31557d;border:1px solid #d8e3ef;padding:10px 15px;border-radius:9px;font-weight:700;font-size:12px}
.logout:hover{background:#edf5ff}
.view{display:none}.view.active{display:block;animation:frameIn .2s ease-out}
@keyframes frameIn{from{opacity:.4;transform:translateY(5px)}to{opacity:1;transform:translateY(0)}}
.mini-overline{display:block;font-size:10px;line-height:1.5;letter-spacing:1.7px;color:#2878d4;font-weight:820;text-transform:uppercase}
.workspace-intro{display:flex;justify-content:space-between;align-items:center;gap:23px;margin-bottom:30px}
.workspace-intro h2{font-size:clamp(25px,2.7vw,37px);margin:5px 0 6px;color:#132b48;line-height:1.25;font-weight:830;letter-spacing:-1.1px}
.workspace-intro p{font-size:13px;line-height:1.7;margin:0;max-width:700px;color:#6b829a}
.section-heading{display:flex;align-items:center;justify-content:space-between;gap:18px;padding-bottom:19px}
.section-heading h2{font-size:21px;letter-spacing:-.55px;font-weight:800;color:#193354;margin:3px 0}
.section-heading p{font-size:12px;color:#71849d;margin:4px 0 0}
.section-counter{font-size:11px;color:#346ca7;background:#eaf3ff;padding:8px 13px;border-radius:100px;font-weight:780}
.text-action{font-size:12px;color:#0e65ce;background:none;border:0;font-weight:800;padding:9px 3px;white-space:nowrap}
.text-action:hover{color:#063e86;text-decoration:underline}
.primary,.secondary,.danger,.hero-cta,.hero-text-link{
display:inline-flex;justify-content:center;align-items:center;gap:9px;
border-radius:10px;min-height:42px;padding:10px 16px;font-size:12px;line-height:1.4;
font-weight:770;text-decoration:none;transition:transform .13s,background .13s,border-color .13s,box-shadow .13s
}
.primary{background:#166df0;border:1px solid #166df0;color:white;box-shadow:0 4px 12px #1673ed21}
.primary:hover{background:#0858cf;border-color:#0858cf;transform:translateY(-1px);filter:none}
.secondary{background:#fff;color:#255380;border:1px solid #cddbeb}
.secondary:hover{background:#eef5ff;border-color:#9fc7f3;transform:translateY(-1px)}
.danger{background:#fff;color:#a42e4b;border:1px solid #ead1d8}
.danger:hover{background:#fff0f3;border-color:#dc9daa}
button:disabled{opacity:.5;transform:none!important}
a:focus-visible,button:focus-visible,summary:focus-visible,input:focus-visible,select:focus-visible{outline:3px solid #57bdff;outline-offset:3px}
.hidden,[hidden]{display:none!important}.sr-only{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0}
.field{font-size:12px;font-weight:740;line-height:1.5;color:#365577;margin-bottom:16px;display:block}
.field input,.field select{
font-size:13px;min-height:43px;display:block;margin-top:6px;width:100%;
padding:10px 13px;color:#193858;background:#fff;border:1px solid #ccdaeb;
border-radius:9px;box-shadow:0 1px 2px #11294b07
}
.field-help{display:block;color:#7a8ba0;font-weight:460;margin:3px 0 5px;font-size:11px;line-height:1.55}
input[type=checkbox]{width:16px;height:16px;accent-color:#1174ea}
.row{display:flex;align-items:center;gap:9px;flex-wrap:wrap}
.inline{display:inline-flex;align-items:center;gap:8px}
.notice{background:#eaf4ff;border:1px solid #bed7f6;color:#235b90;border-radius:12px;padding:13px 17px;margin-bottom:18px}
.notice.bad{background:#fff0f3;border-color:#e6b9c4;color:#9a2542}
.status{display:inline-flex;font-size:10px;color:#06745e;background:#e9fbf4;border:1px solid #a9e4cf;padding:6px 10px;border-radius:99px;font-weight:790}
.status.off{color:#9a6518;background:#fff5e5;border-color:#eed8b2}
.pill{color:#2267a5;background:#e9f3ff;border:1px solid #cfdfef;border-radius:100px;font-size:11px;font-weight:800;padding:6px 10px}
.muted{font-size:12px;color:#687f98}
.panel{background:#fff;border-radius:15px;border:1px solid #e0e8f1;box-shadow:var(--shadow);padding:24px;margin:0 0 19px}
.panel p{line-height:1.7}
.panel h2{font-weight:800;color:#1d3a58;font-size:18px;margin:0 0 6px;letter-spacing:-.4px}
.panel-head{display:flex;justify-content:space-between;align-items:center;gap:12px;margin:0 0 18px;flex-wrap:wrap}
.panel-head h2{margin:0}
.settings-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));column-gap:20px;row-gap:4px}
.settings-grid .field{margin:0 0 18px}
.selected,.review-grid>div{border:1px solid #d8e5f2;border-radius:11px;background:#f8fbff;padding:14px}
.selected{display:flex;align-items:center;justify-content:space-between;gap:15px;flex-wrap:wrap;margin-bottom:14px}
.selected strong{overflow-wrap:anywhere}
.step{display:none}.step.active{display:block}
.step-indicator{display:flex;gap:7px;flex-wrap:wrap;margin:0 0 23px}
.step-indicator span{font-size:11px;padding:10px 13px;border-radius:9px;border:1px solid #dce5f0;background:#f2f6fb;color:#78899d}
.step-indicator span.active{color:#065fc0;background:#e3f0ff;border-color:#abd0ff;font-weight:840}
.wizard-status{padding:13px;border-radius:10px;background:#eff7ff;border:1px solid #c7e0fb;color:#366a9a}
.wizard-status.good{background:#eafaef;border-color:#b2e7ca;color:#23724e}
.wizard-status.bad{background:#fff0f3;border-color:#e8c0cb;color:#9d3451}
.review-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}
.review-grid strong{display:block;font-size:13px;color:#1d4268;overflow-wrap:anywhere}
.review-grid .card-label{font-size:11px;display:block;color:#768a9f;margin-bottom:6px}
.simple-details,.glossary-fold,.onboarding-disclosure,.advanced-update{
border:1px solid #dbe6f1;border-radius:11px;background:#fff;padding:14px 17px;margin-bottom:17px;color:#41688f
}
.simple-details>summary,.glossary-fold>summary,.onboarding-disclosure>summary,.advanced-update>summary{cursor:pointer;font-size:12px;font-weight:770;color:#376ea3}
.simple-details[open]>summary,.glossary-fold[open]>summary,.onboarding-disclosure[open]>summary{margin-bottom:13px}
.glossary-layout{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:13px;color:#657e99;font-size:12px;margin-top:14px}
.glossary-layout strong{color:#245c95}
.plain-guide{font-size:12px;padding:12px 15px;background:#eff7ff;border:1px solid #cae0f8;border-radius:10px;margin-bottom:18px}
.quick-start-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:13px}
.quick-step{background:#f8fbff;border:1px solid #e0eafa;padding:17px;border-radius:12px}
.quick-step h3{font-size:13px}.quick-step p{color:#687f98;font-size:12px;line-height:1.65}
.step-number{display:grid;place-items:center;background:#e4f1ff;border-radius:9px;width:28px;height:28px;font-weight:850;color:#1768cd}
.quick-start-heading h2{font-size:17px}
.quick-start{padding:15px}
.update-ready{background:#fff}

/* Signature overview: strong product identity, not generic boxes */
.command-hero{
display:grid;grid-template-columns:1.2fr .8fr;min-height:335px;
background:#0d2855;border-radius:20px;overflow:hidden;position:relative;
box-shadow:0 16px 42px #0e30502b;margin-bottom:20px;color:#fff
}
.command-hero::before{content:"";position:absolute;width:410px;height:410px;right:-148px;top:-210px;border:1px solid #64b8ff29;border-radius:50%;pointer-events:none}
.command-hero::after{content:"";position:absolute;width:520px;height:520px;right:-183px;top:-258px;border:1px solid #64b8ff20;border-radius:50%;pointer-events:none}
.hero-text{padding:38px 38px 39px;z-index:1}
.hero-overline{display:flex;align-items:center;gap:9px;color:#7dcbfb;font-weight:820;letter-spacing:1.8px;font-size:10px}
.hero-overline-dot{height:8px;width:8px;border-radius:50%;background:var(--aqua);box-shadow:0 0 0 5px #10d7d422}
.hero-text h2{font-weight:850;letter-spacing:-1.9px;font-size:clamp(32px,3.6vw,53px);line-height:1.08;margin:22px 0 14px;color:#fff}
.hero-text h2 em{font-style:normal;color:#4adbf2}
.hero-text p{color:#b1cce9;font-size:13px;line-height:1.75;max-width:520px;margin:0 0 24px}
.hero-actions{display:flex;align-items:center;gap:21px;flex-wrap:wrap}
.hero-cta{background:#16cfd6;border:1px solid #16cfd6;color:#07284a;font-weight:850;box-shadow:0 4px 16px #10bcd32f}
.hero-cta:hover{background:#61e5e6;border-color:#61e5e6;transform:translateY(-2px)}
.hero-text-link{background:transparent;border:0;color:#deedff;padding:10px 0}
.hero-text-link:hover{color:#59e0ff}
.hero-illustration{display:flex;justify-content:center;align-items:center;flex-direction:column;position:relative;padding:20px 16px;z-index:1;background:radial-gradient(circle at center,#1b4c8c4f,transparent 65%)}
.hero-flow-art{width:100%;height:auto;max-width:400px}
.hero-path{color:#2fa2f880}
.hero-path-low{color:#2bdddb88}
.hero-flow-foot{font-size:9px;font-weight:780;color:#88b7e6;letter-spacing:1.2px}
.hero-flow-foot span{color:var(--aqua);padding:0 8px}
.overview-metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));background:#fff;border:1px solid #dde6f2;border-radius:15px;overflow:hidden;margin-bottom:25px;box-shadow:var(--shadow)}
.overview-metric{display:flex;gap:14px;padding:20px 17px;min-width:0;border-right:1px solid #e2e9f1}
.overview-metric:last-child{border-right:0}
.metric-symbol{width:37px;height:37px;min-width:37px;border-radius:11px;background:#eaf3ff;color:#186bd6;display:grid;place-items:center;font-size:20px;font-weight:720}
.overview-metric small{display:block;font-size:10px;font-weight:830;letter-spacing:1px;color:#7189a1}
.overview-metric strong{font-size:24px;font-weight:830;letter-spacing:-.6px;color:#142d4f;display:block;line-height:1.4;font-variant-numeric:tabular-nums;overflow-wrap:anywhere}
.overview-metric>div>span{font-size:11px;color:#8b9bb0}
.overview-columns{display:grid;grid-template-columns:1.08fr .92fr;gap:22px;margin-bottom:22px}
.overview-list-panel{background:#fff;border:1px solid #e0e8f2;border-radius:15px;padding:25px;min-width:0}
.flow-list{border-top:1px solid #e8eef6}
.flow-row,.place-summary{display:flex;align-items:center;gap:14px;padding:16px 0;border-bottom:1px solid #edf1f7}
.flow-row:last-child,.place-summary:last-child{border-bottom:0}
.flow-row-icon,.place-summary-icon{width:38px;height:38px;min-width:38px;display:grid;place-items:center;background:#eaf3ff;color:#156cd8;border-radius:11px;font-size:19px}
.flow-row-content,.place-summary>div:nth-child(2){flex:1;min-width:0}
.flow-row-content strong,.place-summary strong{display:block;font-size:13px;font-weight:790;color:#1c385a;overflow-wrap:anywhere}
.flow-row-content small,.place-summary small{font-size:11px;color:#76899f;display:block;margin-top:2px}
.flow-row-open{border:0;background:#f0f5fc;color:#176bd2;border-radius:9px;width:34px;height:34px;min-width:34px;font-size:18px}
.flow-row-open:hover{background:#dbeafd}
.flow-state{display:inline-flex;align-items:center;gap:7px;font-size:10px;font-weight:830;border-radius:100px;padding:6px 10px;white-space:nowrap}
.flow-state span{width:7px;height:7px;border-radius:50%;background:currentColor}
.flow-state.is-on{background:#e7f9f1;color:#148264}.flow-state.is-off{background:#f1f4f8;color:#71849a}
.place-summary-icon.cloud-symbol{color:#065fc5}
.place-indicator{font-size:10px;font-weight:760;color:#557da6;max-width:112px;text-align:right}
.overview-empty{padding:26px 12px;text-align:center;display:grid;justify-items:center;gap:8px}
.overview-empty>span{font-size:28px;color:#a4c7f1}
.overview-empty strong{font-size:14px;color:#254668}
.overview-empty p{margin:0 0 12px;font-size:12px;color:#79899f}
.overview-empty-compact{padding:18px}
.welcome-track{background:#fff;border:1px solid #dfe7f2;padding:25px;border-radius:16px;margin-bottom:25px}
.welcome-track-intro{margin-bottom:17px}.welcome-track h2{margin:3px 0 5px;font-size:20px;letter-spacing:-.45px}
.welcome-track p{margin:0;color:#778aa0;font-size:12px}
.welcome-track-steps{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:15px}
.welcome-track-steps>div{display:flex;flex-direction:column;align-items:flex-start;border:1px solid #e0e9f4;background:#f8fbff;padding:18px;border-radius:12px}
.track-index{font-size:11px;font-weight:850;color:#1879ec;letter-spacing:1px}
.welcome-track-steps strong{color:#193c63;font-size:13px;margin-top:10px}
.welcome-track-steps p{font-size:12px;line-height:1.6;margin:6px 0 13px;flex:1}
.welcome-track-steps button{border:0;background:transparent;color:#1264c6;font-size:12px;font-weight:810;padding:4px 0}
.welcome-track-steps button:hover{text-decoration:underline}

/* Flow cards: actual controls read left to right, no nested white cards */
.sync-library{display:grid;gap:18px;margin-bottom:22px}
.relationship-card{
display:block!important;background:#fff;border:1px solid #dbe5f1;border-radius:16px;padding:0;
overflow:hidden;box-shadow:0 5px 23px #1439690b
}
.flow-card-main{display:flex;gap:16px;align-items:center;padding:23px 25px 20px}
.flow-card-emblem{width:54px;height:54px;min-width:54px;border-radius:13px;display:grid;place-items:center;background:#e7f2ff;color:#1166d2;font-weight:820;font-size:25px}
.flow-card-identity{flex:1;min-width:0}
.flow-card-identity h3{font-size:19px;font-weight:830;color:#163657;letter-spacing:-.45px;margin:4px 0 2px;overflow-wrap:anywhere}
.flow-card-identity p{font-size:12px;color:#6f8196;margin:0;line-height:1.55}
.flow-card-health{align-self:flex-start}
.flow-card-facts{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));margin:0 24px;padding:16px 0;border-top:1px solid #e8eef6;border-bottom:1px solid #e8eef6}
.flow-card-facts>div{padding:0 18px;border-right:1px solid #e9eef4;min-width:0}
.flow-card-facts>div:first-child{padding-left:0}.flow-card-facts>div:last-child{border-right:0}
.flow-card-facts small{display:block;color:#8a9aab;letter-spacing:1px;font-size:9px;font-weight:830}
.flow-card-facts strong{display:block;margin-top:5px;color:#265073;font-size:12px;overflow-wrap:anywhere}
.flow-card-description{padding:13px 25px 0;margin:0;color:#5e7591;font-size:12px;line-height:1.6}
.flow-card-stats{margin:6px 25px 18px;font-size:11px;color:#8b9caf}
.flow-card-error:not([hidden]){background:#fff2f5;color:#a92e4a;border:1px solid #f0c7d2;margin:10px 24px 16px;padding:13px;border-radius:10px;font-size:12px}
.flow-card-actions{display:flex;align-items:center;gap:10px;flex-wrap:wrap;background:#f8fbff;padding:16px 23px;border-top:1px solid #e5edf6}
.flow-danger{margin-left:auto}
.flow-danger .danger{border-color:#e8d7dc}
.sync-live-progress{margin:14px 24px 18px;padding:14px 16px;background:#edf6ff;border:1px solid #cbdff5;border-radius:11px}
.sync-live-progress progress{width:100%;height:11px;accent-color:#167af2;margin:8px 0}
.sync-live-progress p{margin:0;font-size:11px;color:#66819d}
.progress-topline{display:flex;justify-content:space-between;align-items:center;gap:15px}
.progress-topline small{font-size:10px;color:#6d87a2}
.flow-library-empty{border:1px dashed #b8cae3;border-radius:15px;padding:50px 28px;background:#fff;display:grid;justify-items:center;text-align:center;gap:9px}
.flow-library-empty h3{font-size:21px;font-weight:830;letter-spacing:-.6px;color:#193a60;margin:4px 0}
.flow-library-empty p{font-size:12px;color:#7088a2;line-height:1.7;max-width:400px;margin:0 0 16px}
.empty-symbol{width:63px;height:63px;display:grid;place-items:center;border-radius:16px;background:#eaf3ff;color:#1772dd;font-size:30px;margin-bottom:11px}
.relationship-alert{border:1px solid #e6d6ab;background:#fffaf0;border-radius:12px;padding:17px;margin:0 24px 15px}
.relationship-alert h4{font-size:13px;color:#66491b;margin:0 0 2px}
.relationship-alert p{font-size:12px;color:#896b3f;margin:0}
.alert-title{display:flex;gap:12px;align-items:center}
.alert-title>span{width:32px;min-width:32px;height:32px;border-radius:9px;background:#f8e7b9;color:#8d5e0b;display:grid;place-items:center;font-weight:850;font-size:18px}
.recovery-alert{background:#eef8ff;border-color:#cce5fb}
.recovery-alert .alert-title>span{background:#d7efff;color:#176db3}
.recovery-alert h4{color:#205779}.recovery-alert p{color:#6386a5}
.decision-row{display:flex;align-items:center;justify-content:space-between;gap:15px;border-top:1px solid #ebdfc7;margin-top:13px;padding-top:13px}
.decision-row>div{flex:1;min-width:0}.decision-row strong,.decision-row small{display:block;overflow-wrap:anywhere}
.decision-row small{font-size:11px;color:#816b50}
.decision-row .field{margin:0}.decision-row .field select{margin:0}
.decision-row form{display:flex;gap:8px;flex-wrap:wrap}

/* Storage connections */
.connection-drawer{background:#fff;border:1px solid #c9ddf5;border-top:4px solid #15bccc;border-radius:15px;padding:25px 27px;box-shadow:0 10px 26px #12497e11;margin:0 0 25px}
.drawer-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:25px;margin-bottom:24px}
.drawer-heading h3{font-size:23px;letter-spacing:-.65px;color:#173657;margin:5px 0}
.drawer-heading p{font-size:12px;color:#6a819b;margin:4px 0}
.drawer-footer{display:flex;justify-content:space-between;align-items:center;gap:15px;border-top:1px solid #e1e8f2;padding-top:17px;margin-top:7px}
.drawer-footer p{margin:0;font-size:12px;color:#6e839c}
.connection-list,.activity-table-shell{background:#fff;border:1px solid #dce6f1;border-radius:15px;padding:23px 26px;margin-bottom:23px;box-shadow:var(--shadow)}
.connection-row{display:flex;align-items:center;gap:16px;padding:19px 0;border-top:1px solid #e7edf5}
.connection-row:first-of-type{border-top:0}
.connection-emblem{width:49px;height:49px;min-width:49px;border-radius:12px;background:#eaf3ff;color:#156bce;display:grid;place-items:center}
.connection-copy{flex:1;min-width:0}
.connection-copy strong{font-size:15px;font-weight:790;color:#214466;display:block;overflow-wrap:anywhere}
.connection-copy p,.connection-copy small{margin:4px 0 0;display:block;color:#70869e;font-size:12px;overflow-wrap:anywhere}
.connection-copy small{font-size:11px}
.connection-tools{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.connection-state{background:#eaf7f7;color:#157e83;border-radius:100px;padding:6px 10px;font-size:10px;font-weight:800;max-width:140px;text-align:center}

/* Cloud provider workspaces */
.provider-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:20px;margin-bottom:24px}
.provider-panel{display:flex;flex-direction:column;min-width:0;border:1px solid #dce6f1;border-radius:17px;background:#fff;overflow:hidden;box-shadow:var(--shadow)}
.provider-top{display:flex;gap:14px;align-items:center;padding:25px 23px 22px;border-bottom:1px solid #e8eef6;min-width:0}
.provider-icon{display:grid;place-items:center;width:51px;height:51px;min-width:51px;border-radius:13px;font-weight:850;font-size:26px}
.google-icon{background:#eef4ff;color:#2c71d9}.dropbox-icon{background:#e7f3ff;color:#1777d7}
.provider-top>div{flex:1;min-width:0}
.provider-top h3{font-size:21px;font-weight:830;letter-spacing:-.65px;color:#1a3c5e;margin:4px 0}
.provider-top p{color:#7d8ea4;font-size:11px;margin:0}
.provider-count{font-size:10px;font-weight:780;color:#5b7fa5;background:#eef5fd;padding:6px 9px;border-radius:99px;white-space:nowrap}
.provider-list{flex:1;padding:0 23px}
.provider-account{display:flex;align-items:center;justify-content:space-between;gap:10px;padding:18px 0;border-bottom:1px solid #e8eef6}
.provider-account>div:first-child{min-width:0}
.provider-account strong,.provider-account small{display:block}
.provider-account strong{font-size:13px;font-weight:770;color:#224566;overflow-wrap:anywhere}
.provider-account small{font-size:11px;color:#748ca4;margin-top:4px}
.provider-buttons{display:flex;gap:8px;flex-wrap:wrap}
.provider-empty{padding:38px 0;color:#8294a9;font-size:12px}
.provider-add{margin:15px 23px 22px;border:1px dashed #b6d1ef;background:#f6faff;color:#1465c1;padding:13px 16px;border-radius:10px;display:flex;justify-content:space-between;gap:8px;font-size:12px;font-weight:800}
.provider-add:hover{background:#eaf3ff;border-style:solid}

/* Activity ledger */
.activity-scoreboard{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px;margin-bottom:23px}
.activity-score{padding:20px 22px;border:1px solid #e0e8f1;border-radius:13px;background:#fff}
.activity-score-highlight{background:#103360;border-color:#103360;box-shadow:0 10px 25px #10336024}
.activity-score small{font-size:10px;display:block;font-weight:830;letter-spacing:1.1px;color:#758aa0}
.activity-score strong{font-size:34px;font-weight:870;letter-spacing:-1px;display:block;margin:7px 0 1px;color:#1a3d62;font-variant-numeric:tabular-nums;overflow-wrap:anywhere}
.activity-score span{font-size:11px;color:#8999ab}
.activity-score-highlight small{color:#83bdea}.activity-score-highlight strong{color:#fff}.activity-score-highlight span{color:#bdd5ee}
.activity-row{padding:21px 0;border-top:1px solid #e5edf6}
.activity-row-head{display:flex;justify-content:space-between;align-items:center;gap:15px;margin-bottom:16px}
.activity-row-head strong,.activity-row-head small{display:block}
.activity-row-head strong{font-size:15px;color:#1a3e66;font-weight:810}
.activity-row-head small{font-size:11px;margin-top:3px;color:#7189a1}
.activity-state{padding:6px 11px;background:#ecf6ff;color:#245e98;font-size:11px;font-weight:750;border-radius:100px;white-space:nowrap}
.activity-state.off{background:#fff4ed;color:#a0612d}
.activity-row-numbers{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px}
.activity-row-numbers>div{background:#f7faff;padding:12px;border-radius:9px}
.activity-row-numbers small{font-size:9px;font-weight:820;letter-spacing:1px;color:#879aad;display:block}
.activity-row-numbers strong{font-size:20px;line-height:1.5;font-weight:830;color:#224468}
.activity-row-detail{font-size:11px;color:#758aa2;margin:13px 0 0;overflow-wrap:anywhere}

/* Updates and advanced workspace, with clean idle state */
.update-workspace{display:grid;grid-template-columns:minmax(0,1.45fr) minmax(280px,.8fr);gap:19px;align-items:start}
.update-workspace>div{min-width:0}
.update-hero{background:#eaf3ff;border:1px solid #cfdef5;border-radius:13px;padding:21px;display:flex;gap:20px;align-items:center;justify-content:space-between;margin-bottom:17px}
.update-hero strong{display:block;font-size:25px;font-weight:830;color:#183e69}
.update-hero p{font-size:12px;color:#64809f;margin:4px 0}
.update-revisions{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:11px;margin:15px 0 19px}
.update-revisions>div{padding:13px;background:#f7faff;border-radius:10px;border:1px solid #e3ebf6;min-width:0}
.update-revisions small{font-size:10px;letter-spacing:1px;color:#8597ad;text-transform:uppercase;display:block}
.update-revisions strong{display:block;font-size:15px;color:#1f456d;overflow-wrap:anywhere}
.update-progress-panel{border-radius:11px;background:#f6faff;border:1px solid #d8e7f8;padding:18px;margin-top:12px}
.update-progress-panel progress{width:100%;height:16px;accent-color:#1875f6;margin:15px 0 8px}
.update-progress-meta{display:flex;justify-content:space-between;gap:15px;align-items:center}
.update-progress-meta strong{color:#23507c}
.update-connection{border-radius:10px;padding:10px 12px;font-size:11px;color:#3e7a9d;background:#eaf6fc;border:1px solid #cce9f6}
.update-connection.lost{color:#9a661b;border-color:#f2d9aa;background:#fff8ed}
.update-measures{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:9px;margin:14px 0}
.update-measures>div{padding:12px;border-radius:9px;background:#fff;border:1px solid #e1e9f3;min-width:0}
.update-measures small{display:block;font-size:10px;color:#879caf}
.update-measures strong{display:block;color:#24517b;font-size:12px;overflow-wrap:anywhere}
.update-steps{display:flex;list-style:none;gap:7px;flex-wrap:wrap;padding:0;margin:15px 0}
.update-steps li{padding:6px 9px;border:1px solid #dce5f2;background:#fff;color:#758ca5;font-size:11px;border-radius:7px}
.update-steps li.active{color:#075ebc;border-color:#b7d6fa;background:#e8f3ff}
.update-steps li.done{color:#137953;border-color:#bce3d0;background:#eafaef}
.update-steps li.failed{color:#a7354e;border-color:#edc5d0;background:#fff1f3}
.update-explainer{display:grid;gap:15px}
.update-explainer>div{display:flex;align-items:flex-start;gap:13px}
.update-explainer>div>span{height:31px;width:31px;min-width:31px;display:grid;place-items:center;background:#eaf3ff;color:#146fd9;border-radius:9px;font-weight:840}
.update-explainer p{margin:0}
.update-explainer strong,.update-explainer small{display:block}
.update-explainer strong{font-size:12px;color:#1d456d}
.update-explainer small{font-size:11px;color:#8397aa;margin-top:4px}
.update-timeline{list-style:none;margin:0;padding:0}
.update-timeline li{border-left:2px solid #d8e7f6;padding:0 0 17px 17px;margin-left:9px;position:relative;color:#698299;font-size:11px}
.update-timeline li:before{content:"";position:absolute;width:8px;height:8px;left:-5px;top:6px;border-radius:50%;background:#9fbbd6}
.update-timeline li.active:before{background:#1573f0;box-shadow:0 0 0 4px #e8f3ff}
.update-timeline li.failed:before{background:#b52f50}
.update-timeline strong,.update-timeline small{display:block}
.update-timeline strong{color:#24547f;font-size:12px}
.update-timeline small{font-size:10px;margin-top:3px}
.update-live-banner{border:1px solid #b7e3f2;border-radius:11px;background:#e9f9ff;padding:12px 15px;display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:18px}
.update-live-banner strong{font-size:12px;color:#146c92}
.update-live-banner small{font-size:10px;color:#678bab}
.update-alert{color:#a62e4d;background:#fff2f4;border:1px solid #ecc8d2;padding:12px;border-radius:9px}
.advanced-update>summary{font-size:12px}.advanced-update-body{border-top:1px solid #e3eaf3;margin-top:14px;padding-top:17px}

/* Existing folder picker and settings retain all functional controls */
.picker{background:#0715288a}
.picker-box{background:#fff;border:1px solid #dce7f3;border-radius:16px;box-shadow:0 28px 90px #0819324d}
.folder-row{border-radius:8px}
.folder-row:hover{background:#eaf3ff}
.settings-grid .field input{max-width:100%}
.auth-layout{max-width:620px!important;grid-template-columns:1fr!important;margin:9vh auto;min-height:0}
.auth-layout .content{padding:30px}
.auth-layout .panel{box-shadow:0 15px 45px #122c4c18;padding:34px}
.auth-brand{max-width:255px;height:auto}

/* Responsive design: functions first, not a squeezed desktop */
@media(max-width:1300px){
 .app-layout{grid-template-columns:244px minmax(0,1fr)}
 .sidebar{padding:22px 13px}
 .overview-columns{grid-template-columns:1fr 1fr}
 .overview-metric{padding:15px 12px;gap:9px}
 .metric-symbol{width:32px;min-width:32px;height:32px}
 .overview-metric strong{font-size:22px}
}
@media(max-width:1050px){
 .app-layout{grid-template-columns:90px minmax(0,1fr)}
 .sidebar{padding:17px 11px}
 .sidebar-logo{padding:7px;height:54px;border-radius:11px}
 .sidebar-logo img{width:66px;height:49px;object-fit:cover;object-position:left}
 .sidebar-edition,.nav-section-label,.nav-copy,.sidebar-presence,.sidebar-version{display:none}
 .app-layout .nav button{justify-content:center;padding:8px;min-height:48px}
 .nav-icon{width:35px;min-width:35px}
 .sidebar-navigation{gap:19px}
 .overview-columns,.provider-grid{grid-template-columns:1fr}
 .command-hero{grid-template-columns:1.1fr .9fr}
 .hero-text{padding:30px}
 .hero-text h2{font-size:37px}
 .update-workspace{grid-template-columns:1fr}
}
@media(max-width:760px){
 .app-layout{display:block!important}
 .sidebar{height:auto;min-height:0;position:relative;overflow:visible;padding:13px 15px;gap:12px}
 .sidebar-brand{display:flex;align-items:center;gap:14px;padding:0}
 .sidebar-logo{height:47px;width:139px;padding:4px 10px}
 .sidebar-logo img{width:116px;height:43px;object-fit:contain}
 .sidebar-edition{display:flex;margin:0;font-size:9px}
 .sidebar-edition-line{display:none}
 .sidebar-navigation{display:block;overflow-x:auto;scrollbar-width:thin;margin:0}
 .nav-group{display:inline-flex;gap:2px}
 .nav-section-label{display:none}
 .app-layout .nav{display:flex;gap:5px}
 .app-layout .nav button{width:45px;min-width:45px;height:43px;min-height:43px;padding:5px}
 .nav-copy{display:none}
 .nav-icon{width:28px;min-width:28px;height:28px;background:transparent}
 .sidebar-bottom{display:none}
 .content{padding:24px 17px 60px}
 .topline{margin-bottom:24px;padding-bottom:18px}
 .header-version{display:none}
 .command-hero{grid-template-columns:1fr}
 .hero-text{padding:30px 25px 11px}
 .hero-text h2{font-size:36px}
 .hero-illustration{max-height:190px;padding:0 15px 10px}
 .hero-flow-art{height:155px}
 .overview-metrics{grid-template-columns:repeat(2,minmax(0,1fr))}
 .overview-metric:nth-child(2){border-right:0}
 .overview-metric:nth-child(-n+2){border-bottom:1px solid #e2e9f1}
 .welcome-track-steps{grid-template-columns:1fr}
 .workspace-intro{align-items:flex-start;flex-direction:column}
 .activity-scoreboard{grid-template-columns:repeat(2,minmax(0,1fr))}
}
@media(max-width:530px){
 .topline h1{font-size:25px}
 .content{padding:21px 13px 55px}
 .workspace-intro h2{font-size:27px}
 .overview-list-panel,.connection-list,.activity-table-shell{padding:18px}
 .overview-metric{padding:14px 11px}
 .overview-metric strong{font-size:21px}
 .overview-metric small{font-size:9px}
 .overview-metric>div>span{font-size:10px}
 .flow-row,.place-summary{gap:9px}
 .flow-card-main{padding:18px;gap:12px;flex-wrap:wrap}
 .flow-card-emblem{width:41px;height:41px;min-width:41px;font-size:19px}
 .flow-card-identity{min-width:125px}
 .flow-card-health{margin-left:auto}
 .flow-card-facts{margin:0 16px;gap:9px;grid-template-columns:1fr}
 .flow-card-facts>div{border-right:0;border-bottom:1px solid #e8eef6;padding:7px 0!important}
 .flow-card-facts>div:last-child{border-bottom:0}
 .flow-card-actions{padding:14px 16px}
 .flow-card-description{padding:13px 17px 0}
 .flow-card-stats{margin:7px 17px 14px}
 .flow-danger{margin-left:0}
 .connection-row,.provider-account,.decision-row{flex-wrap:wrap}
 .connection-tools,.provider-buttons{width:100%}
 .settings-grid,.review-grid,.activity-row-numbers,.update-revisions,.update-measures{grid-template-columns:1fr 1fr}
 .provider-grid{gap:14px}
 .provider-top{flex-wrap:wrap}
 .drawer-heading,.drawer-footer{flex-direction:column;align-items:flex-start}
 .glossary-layout{grid-template-columns:1fr}
 .hero-actions{gap:12px}
}
@media(prefers-reduced-motion:reduce){*,*:before,*:after{animation:none!important;transition:none!important}}

/* Product-specific sync builder: its own composition, not a generic modal card */
.sync-builder{
 background:#fff;border:1px solid #d6e4f4;border-top:5px solid #16ccd0;
 border-radius:16px;margin:27px 0 30px;padding:0;overflow:hidden;
 box-shadow:0 15px 43px #103b6d17
}
.builder-heading{display:flex;justify-content:space-between;align-items:flex-start;gap:18px;background:#102e59;color:#fff;padding:27px 31px}
.builder-heading .mini-overline{color:#4cdeea}
.builder-heading h2{font-size:27px;letter-spacing:-.8px;color:#fff;font-weight:840;margin:5px 0}
.builder-heading p{font-size:12px;color:#b3cee9;margin:5px 0 0}
.builder-heading .secondary{background:#1c487d;border:1px solid #3e6799;color:#e3f2ff}
.builder-heading .secondary:hover{background:#285a96}
.sync-builder>.step-indicator{background:#f1f7ff;padding:16px 28px;margin:0;border-bottom:1px solid #dfeaf7;gap:10px}
.sync-builder>.step-indicator span{background:transparent;border:0;border-radius:0;padding:8px 12px;color:#7b93ad;font-size:11px;font-weight:750}
.sync-builder>.step-indicator span.active{color:#0e66cb;background:#fff;border-radius:8px;box-shadow:0 3px 9px #1954a211}
.sync-builder>form{padding:29px 31px 33px;max-width:930px}
.sync-builder .step>h2{font-size:23px;color:#1b3c61;letter-spacing:-.6px;font-weight:830;margin:0 0 5px}
.sync-builder .step>p.muted{color:#7189a0;margin:0 0 22px;line-height:1.6;font-size:12px}
.sync-builder .step .field{max-width:790px}
.sync-builder .wizard-status{margin:20px 0}
.sync-builder .selected{background:#f6fbff}
.sync-builder .row{margin-top:13px}
.sync-builder .row .primary{min-height:43px}
.sync-builder #nasTargets{border:1px solid #e3ebf5;border-radius:13px;padding:13px;background:#f9fcff}
.sync-builder .review-grid{margin:13px 0 18px}
.sync-builder .step-indicator{margin:0 0 24px}
.release-panel{border-top:4px solid #12cad4}
.release-panel-heading{display:flex;align-items:center;justify-content:space-between;gap:13px;margin-bottom:18px}
.release-well{padding:24px;border-radius:13px;background:#0f305e;color:#fff;display:flex;justify-content:space-between;align-items:center;gap:15px;margin:0 0 15px}
.release-well>div{min-width:0}
.release-well small{font-size:10px;letter-spacing:1.2px;font-weight:840;color:#81bdec;display:block}
.release-well strong{font-size:32px;letter-spacing:-.7px;color:#fff;font-weight:850;display:block;margin:3px 0}
.release-well p{color:#aacdeb;font-size:12px;margin:0}
.release-well .secondary{border-color:#36669a;background:#1e4c81;color:#fff;white-space:nowrap}
.release-well .secondary:hover{background:#2467aa}
.release-actions{display:flex;align-items:center;flex-wrap:wrap;gap:11px}
.release-note{font-size:11px;color:#8797ab;margin:17px 0 2px;line-height:1.6}
.update-progress-panel{margin-top:0}
#view-settings .panel{max-width:1020px}
#view-settings .panel:first-of-type{border-top:4px solid #16ccd0}
@media(max-width:680px){
 .sync-builder>.step-indicator{padding:12px}
 .sync-builder>.step-indicator span{font-size:10px;padding:6px 8px}
 .sync-builder>form{padding:18px}
 .builder-heading{padding:22px 18px;flex-wrap:wrap}
 .builder-heading h2{font-size:23px}
 .release-well{flex-direction:column;align-items:flex-start}
 .release-well strong{font-size:26px}
}
`
