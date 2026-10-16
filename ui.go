package main

// newUI is a complete replacement for the original single-page dashboard.
var newUI = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Syncria — Your files, in sync</title>
<style>
/* Functional layout primitives. Product styling is centralized in ui_theme.go. */
*{box-sizing:border-box}
body{margin:0}
button,input,select,textarea{font:inherit}
button{cursor:pointer}
button:disabled{cursor:not-allowed}
a{color:inherit}
main{min-height:100vh}
.view,.step{display:none}
.view.active,.step.active{display:block}
.hidden,[hidden]{display:none!important}
.row{display:flex;gap:10px;align-items:center;flex-wrap:wrap}
form.inline{display:inline}
.stats{display:grid;gap:12px}
.split{display:grid;grid-template-columns:1fr 1fr;gap:17px}
.entry{display:flex;gap:15px;align-items:center;justify-content:space-between}
.empty{padding:25px;text-align:center}
.empty strong{display:block}
.field{display:block}
.picker{position:fixed;inset:0;background:#0a183a9c;z-index:60;display:none;align-items:center;justify-content:center;padding:16px}
.picker.open{display:flex}
.picker-box{width:min(740px,100%);max-height:88vh;display:flex;flex-direction:column;background:white}
.picker-head{padding:20px;border-bottom:1px solid #dde8f3}
.picker-head h2{margin:0}
.picker-path{font-size:12px;color:#6b82a0;margin:8px 0 0;overflow-wrap:anywhere}
.picker-actions{padding:12px 20px;border-bottom:1px solid #dde8f3}
.folder-list{padding:11px 15px;min-height:150px;overflow-y:auto}
.folder-row{display:block;width:100%;text-align:left;background:none;border:0;padding:12px}
.picker-footer{padding:15px 20px;display:flex;align-items:center;justify-content:space-between;gap:12px;border-top:1px solid #dde8f3}
.nas-target{padding:13px;border:1px solid #dce8f4;border-radius:10px;background:#f9fcff}
.nas-target label{display:flex;gap:10px;align-items:center}
.nas-target small{color:#7c91a9}
.nas-target button{white-space:nowrap}
.progress-topline{display:flex;justify-content:space-between;gap:12px}
code{overflow-wrap:anywhere}

.live-sync-workspace{margin-bottom:18px}
.active-sync-list{display:grid;gap:14px}
.active-sync-card{padding:17px;border:1px solid #d8e5f2;border-radius:12px;background:#fff;min-width:0}
.active-sync-heading{display:flex;justify-content:space-between;align-items:center;gap:15px;flex-wrap:wrap}
.active-sync-title{display:grid;gap:3px;min-width:0}
.active-sync-title>strong{font-size:16px;color:#102c4d}
.active-sync-title>small{color:#55728f}
.active-sync-current{font-size:13px;overflow-wrap:anywhere;margin:13px 0 8px}
.active-sync-stats{display:flex;gap:10px 20px;flex-wrap:wrap;font-size:12px;color:#41617e;margin-bottom:12px}
.active-sync-plan-label{font-weight:700;font-size:13px;margin:14px 0 8px}
.active-sync-file-list{display:grid;gap:0;border:1px solid #dbe5ef;border-radius:9px;max-height:360px;overflow:auto}
.active-sync-file{padding:9px 12px;border-bottom:1px solid #edf2f7;display:flex;align-items:center;gap:10px;justify-content:space-between;min-width:0}
.active-sync-file:last-child{border-bottom:none}
.active-sync-file-main{display:flex;gap:10px;min-width:0;align-items:center}
.active-sync-file-action{min-width:72px;text-transform:capitalize;font-size:11px;color:#31658c}
.active-sync-file-path{font-size:12px;overflow-wrap:anywhere;min-width:0}
.active-sync-file-state{color:#5e7992;white-space:nowrap;flex-shrink:0}
.active-sync-file.is-active{background:#edf8ff}
.active-sync-file.is-done{opacity:.65}
@media(max-width:700px){.active-sync-file{align-items:flex-start;flex-direction:column}.active-sync-file-state{padding-left:0}}

.preview-filter-bar{display:flex;align-items:center;justify-content:flex-end;gap:8px;margin:12px 0}
.preview-filter-bar label{font-size:12px;font-weight:600}
.preview-filter-bar select{margin-left:8px;min-height:34px;padding:5px 10px;border:1px solid #d6e4ef;border-radius:7px;background:white}
.preview-table tr[hidden]{display:none}
.active-sync-file-details{min-width:0;display:flex;flex-direction:column;gap:5px}
.active-sync-file-reason{font-size:11px;color:#57738c;white-space:normal;overflow-wrap:anywhere}
.active-sync-file-main{align-items:flex-start}
.active-sync-file-state{align-self:center}
.flow-card-dismiss{margin:0 0 12px auto;display:block;width:max-content;max-width:100%;font-size:12px}
.flow-card-dismiss[hidden]{display:none}

.sync-title-link{font:inherit;font-weight:inherit;color:inherit;background:none;border:0;padding:0;text-align:left;cursor:pointer}
.sync-title-link:hover{text-decoration:underline}
.sync-detail-back{display:flex;justify-content:space-between;gap:12px;margin-bottom:16px}
.sync-details-shell{min-width:0}
.sync-detail-live{background:#eff7fc;border:1px solid #d2e6f5;border-radius:10px;padding:12px;font-size:13px;color:#264e6a;margin-bottom:18px;overflow-wrap:anywhere}
.sync-detail-columns{display:grid;grid-template-columns:minmax(240px,340px) minmax(0,1fr);gap:18px}
.sync-detail-run-column,.sync-detail-files-column{min-width:0}
.sync-run-list,.sync-run-files{border:1px solid #dce7ee;border-radius:10px;min-height:75px;max-height:540px;overflow-y:auto}
.sync-history-run{display:flex;flex-direction:column;text-align:left;align-items:flex-start;width:100%;background:white;border:0;border-bottom:1px solid #e6edf3;padding:12px;cursor:pointer;gap:5px;color:inherit}
.sync-history-run:hover,.sync-history-run.selected{background:#edf6fd}
.sync-history-run-status{text-transform:capitalize;font-weight:600;color:#416783}
.sync-history-run-stats,.sync-history-run-count{font-size:12px;color:#5b7487}
.sync-history-file{padding:12px;border-bottom:1px solid #e6edf3;display:flex;flex-direction:column;gap:6px;overflow-wrap:anywhere}
.sync-history-file:last-child{border:0}
.sync-history-file-head{display:flex;justify-content:space-between;gap:10px;flex-wrap:wrap;font-size:12px}
.sync-history-file-head strong{text-transform:capitalize}
.sync-history-file-head span{color:#54738e}
.sync-history-file-path{font-size:13px;font-weight:600}
.sync-history-file small{font-size:12px;color:#496d86}
.sync-detail-pager{display:flex;justify-content:flex-start;margin-top:12px}
@media(max-width:850px){.sync-detail-columns{grid-template-columns:minmax(0,1fr)}.sync-run-list{max-height:320px}}
` + v051Theme + syncriaDesignCSS + `
.computer-layout{display:grid;grid-template-columns:1fr 1fr;gap:18px;align-items:start;margin-bottom:18px}
.computer-device-list,.computer-transfer-list{display:grid;gap:10px}
.computer-device{display:flex;justify-content:space-between;align-items:center;gap:12px;padding:13px 14px;border:1px solid #dce7f0;border-radius:10px;background:#fff}
.computer-device strong,.computer-device small{display:block;overflow-wrap:anywhere}
.computer-device small{color:#658095;margin-top:4px;font-size:12px}
.computer-device .secondary{flex-shrink:0}
.computer-online{color:#237556}.computer-offline{color:#a5483b}
.computer-help{background:#f3f8fc;padding:15px;border:1px solid #dce9f2;border-radius:10px;margin-bottom:16px;display:grid;gap:8px}
.computer-help strong{font-size:12px}.computer-help code{display:block;white-space:pre-wrap;overflow-wrap:anywhere;font-size:12px}
.computer-help small{color:#536e84;line-height:1.5}
.computer-pair-form{display:grid;gap:10px}
.computer-explorer{display:grid;grid-template-columns:minmax(0,1fr) minmax(0,1fr);gap:18px}
.computer-explorer-pane{min-width:0;border:1px solid #dce7f0;border-radius:12px;padding:16px;background:#f9fcfe}
.computer-explorer-pane h3{margin:0 0 14px}
.computer-path{display:flex;align-items:center;gap:10px;padding:10px 0;min-width:0}
.computer-path span{font-size:12px;overflow-wrap:anywhere;min-width:0}
.computer-file-list{height:320px;overflow-y:auto;border:1px solid #dce7f0;background:#fff;border-radius:9px;padding:6px}
.computer-file-list p{padding:8px}
.computer-file{display:flex;align-items:center;justify-content:space-between;width:100%;gap:10px;padding:11px;border:0;border-bottom:1px solid #edf1f6;background:transparent;text-align:left;color:inherit;cursor:pointer}
.computer-file:hover,.computer-file.is-selected{background:#e9f4fd}
.computer-file strong{font-size:13px;overflow-wrap:anywhere}
.computer-file small{font-size:11px;white-space:nowrap;color:#658095}
.computer-file.is-selected{outline:2px solid #6eb6e7;outline-offset:-2px}
.computer-transfer-actions{display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap;border-top:1px solid #e6edf3;margin-top:18px;padding-top:16px}
.computer-transfer-actions strong,.computer-transfer-actions small{display:block}
.computer-transfer-actions small{font-size:12px;color:#5a7489;margin-top:5px}
.computer-transfer-record{padding:12px;border:1px solid #e0eaf1;border-radius:10px;background:#fff;display:grid;gap:5px}
.computer-transfer-record strong{overflow-wrap:anywhere;font-size:13px}.computer-transfer-record small{font-size:12px;color:#57738a}
@media(max-width:950px){.computer-layout,.computer-explorer{grid-template-columns:minmax(0,1fr)}.computer-file-list{height:250px}}
</style></head><body>
<main class="{{if and .Configured .Logged}}app-layout{{else}}auth-layout{{end}}">
{{if and .Configured .Logged}}
<aside class="sidebar" aria-label="Syncria navigation">
  <div class="sidebar-brand"><div class="sidebar-logo"><img src="/assets/syncria-logo.svg" alt="Syncria" width="196" height="70"></div><div class="sidebar-edition"><span class="pulse-dot"></span> Syncria <span class="sidebar-edition-line"></span></div></div>
  <div class="sidebar-navigation">
    <div class="nav-group"><p class="nav-section-label">WORKSPACE</p><nav class="nav" aria-label="Workspace">
      <button type="button" data-view="overview" class="active" aria-label="Overview"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="7" height="7" rx="1.6"/><rect x="14" y="3" width="7" height="7" rx="1.6"/><rect x="3" y="14" width="7" height="7" rx="1.6"/><rect x="14" y="14" width="7" height="7" rx="1.6"/></svg></span><span class="nav-copy"><strong>Overview</strong><small>Your files at a glance</small></span></button>
      <button type="button" data-view="sync" class="" aria-label="My Syncs"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 7h15m0 0-4-4m4 4-4 4M20 17H5m0 0 4-4m-4 4 4 4"/></svg></span><span class="nav-copy"><strong>My Syncs</strong><small>Manage syncs</small></span></button>
      <button type="button" data-view="activity" class="" aria-label="Activity"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 17h4l3-10 4 13 3-7h4"/></svg></span><span class="nav-copy"><strong>Activity</strong><small>Transfers and results</small></span></button>
    </nav></div>
    <div class="nav-group"><p class="nav-section-label">CONNECTIONS</p><nav class="nav" aria-label="Connections">
      <button type="button" data-view="storage" class="" aria-label="My NAS Boxes"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="8" rx="2"/><rect x="3" y="13" width="18" height="8" rx="2"/><path d="M7 7h.01M7 17h.01M11 7h6M11 17h6"/></svg></span><span class="nav-copy"><strong>My NAS Boxes</strong><small>Network storage</small></span></button>
      <button type="button" data-view="computers" aria-label="My Computers"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="4" width="18" height="13" rx="2"/><path d="M8 21h8m-4-4v4"/></svg></span><span class="nav-copy"><strong>My Computers</strong><small>Direct file transfers</small></span></button>
      <button type="button" data-view="accounts" class="" aria-label="Cloud Accounts"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 16.5a4 4 0 0 1 1-7.7 7.5 7.5 0 0 1 14.4 1.2A4.5 4.5 0 0 1 18 19H6.5"/><path d="M8 16h8m-4-4v8"/></svg></span><span class="nav-copy"><strong>Cloud Accounts</strong><small>Google & Dropbox</small></span></button>
    </nav></div>
    <div class="nav-group"><p class="nav-section-label">SYSTEM</p><nav class="nav" aria-label="System">
      <button type="button" data-view="updates" class="" aria-label="App Updates"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4.8 9a7.5 7.5 0 0 1 13.4-3L21 9m0-6v6h-6M19.2 15A7.5 7.5 0 0 1 5.8 18L3 15m0 6v-6h6"/></svg></span><span class="nav-copy"><strong>App Updates</strong><small>Versions and installs</small></span><span id="updateNavIndicator" class="update-nav-indicator hidden" aria-label="Update available">New</span></button>
      <button type="button" data-view="settings" class="" aria-label="Settings"><span class="nav-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 3 14 5l3-1 1 3 3 2-2 3 2 3-3 2-1 3-3-1-2 2-2-2-3 1-1-3-3-2 2-3-2-3 3-2 1-3 3 1z"/><circle cx="12" cy="12" r="3"/></svg></span><span class="nav-copy"><strong>Settings</strong><small>Advanced configuration</small></span></button>
    </nav></div>
  </div>
  <div class="sidebar-bottom"><div class="sidebar-presence"><span class="sidebar-presence-icon"><svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 7h15m0 0-4-4m4 4-4 4M20 17H5m0 0 4-4m-4 4 4 4"/></svg></span><span><strong>Syncria</strong><small>Your files stay yours.</small></span></div><div class="sidebar-version">VERSION <strong><span id="sidebarVersion">{{.Version}}</span></strong></div></div>
</aside>
{{end}}
<div class="content{{if not .Logged}} auth-content{{end}}">
{{if and .Configured .Logged}}<div class="live-system-bar" id="liveSystemBar" data-live-state="idle" role="status" aria-live="off"><span class="live-system-dot" id="liveSystemDot">○</span><span id="liveSystemDetail">Connecting to live sync activity…</span><button type="button" onclick="navigate('activity')">View activity →</button></div>
<header class="topline"><div class="page-heading"><div class="eyebrow"><span class="breadcrumb-home">SYNC WORKSPACE</span><span class="breadcrumb-divider">/</span><span id="pageBreadcrumb">{{if not .Configured}}GETTING STARTED{{else if not .Logged}}WELCOME BACK{{else}}OVERVIEW{{end}}</span></div><h1 id="pageTitle">{{if not .Configured}}Welcome to Syncria{{else if not .Logged}}Sign in{{else}}Your overview{{end}}</h1></div><div class="header-actions">{{if .Logged}}<span class="header-version">Syncria {{.Version}}</span><form method="post" action="/action"><input type="hidden" name="action" value="logout"><button class="logout" type="submit">Sign out ↗</button></form>{{end}}</div></header>{{end}}
{{if .Message}}<div class="notice action-feedback" id="actionFeedback" role="status">{{.Message}}<button type="button" class="feedback-dismiss" onclick="this.parentElement.remove()" aria-label="Dismiss notification">×</button></div>{{end}}{{if and .Configured .Logged}}<div id="updateLiveBanner" class="update-live-banner hidden" role="status"><div><strong>Update running in Syncria</strong><div><small id="updateLiveBannerStatus">Checking the updater status…</small></div></div><button type="button" class="secondary" onclick="navigate('updates')">View exact progress →</button></div>{{end}}{{if and .Configured .Logged}}<div id="sessionExpiredNotice" class="notice bad hidden" role="alert">Your Syncria session has expired. <a href="/">Sign in again</a> to continue.</div>{{end}}
{{if not .Configured}}<section class="panel auth-card"><div class="auth-identity"><img class="auth-brand" src="/assets/syncria-logo.svg" alt="Syncria"></div><div class="auth-intro"><span class="auth-eyebrow">GETTING STARTED</span><h1 class="auth-title">Secure your workspace</h1><p>Create an administrator password to protect your syncs, connected accounts, and activity.</p></div><form class="auth-form" method="post" action="/action"><input type="hidden" name="action" value="setup"><label class="field">Administrator password<input type="password" name="password" minlength="12" maxlength="1024" required autocomplete="new-password"></label><p class="auth-help">Use at least 12 characters.</p><button class="primary auth-submit" type="submit">Secure workspace</button></form></section>
{{else if not .Logged}}<section class="panel auth-card"><div class="auth-identity"><img class="auth-brand" src="/assets/syncria-logo.svg" alt="Syncria"></div><div class="auth-intro"><span class="auth-eyebrow">WELCOME BACK</span><h1 class="auth-title">Sign in to Syncria</h1><p>Enter your administrator password to access your synchronization workspace.</p></div><form class="auth-form" method="post" action="/action"><input type="hidden" name="action" value="login"><label class="field">Administrator password<input type="password" name="password" required autocomplete="current-password" autofocus></label><label class="auth-remember"><input type="checkbox" name="remember" value="on"><span><strong>Remember me</strong><small>Keep me signed in on this device for 30 days.</small></span></label><button class="primary auth-submit" type="submit">Sign in</button></form><p class="auth-help auth-footer">Without Remember me, your browser session lasts until you close the browser (up to 7 days).</p></section>
{{else}}
<section class="view active" id="view-overview">
<div class="overview-toolbar"><div><strong>Workspace</strong><p>Manage your connected folders, transfers, and storage in one place.</p></div><button class="primary" type="button" onclick="startWizard()">+ Create sync</button></div>
<div class="overview-metrics grid gap-3" aria-label="Your current Syncria numbers">
 <div class="overview-metric"><span class="metric-symbol">↗</span><div><small>YOUR SYNCS</small><strong>{{len .Jobs}}</strong><span>configured connections</span></div></div>
 <div class="overview-metric"><span class="metric-symbol">▦</span><div><small>FILES TRACKED</small><strong id="metricFiles">{{.Metrics.TrackedFiles}}</strong><span>across your syncs</span></div></div>
 <div class="overview-metric"><span class="metric-symbol">⇅</span><div><small>DATA MOVED</small><strong id="metricBandwidth">—</strong><span>total transferred</span></div></div>
 <div class="overview-metric"><span class="metric-symbol">✓</span><div><small>SYNC RUNS</small><strong id="metricRuns">{{.Metrics.Runs}}</strong><span><span id="metricFailures">{{.Metrics.Failures}}</span> with errors</span></div></div>
</div>
{{if not .Jobs}}
<div class="welcome-track"><div class="welcome-track-intro"><span class="mini-overline">Getting started</span><h2>Set up your first sync</h2><p>New here? Start with these three steps. You only need to set this up once.</p></div>
<div class="welcome-track-steps">
 <div><span class="track-index">01</span><strong>Connect your cloud</strong><p>Google Drive or Dropbox. Skip this if you're using two NAS boxes.</p><button type="button" onclick="startAccountWizard()">Cloud accounts →</button></div>
 <div><span class="track-index">02</span><strong>Add a NAS box</strong><p>Tell Syncria where your storage box lives.</p><button type="button" onclick="toggleDirectNAS(true)">Storage →</button></div>
 <div><span class="track-index">03</span><strong>Choose your folders</strong><p>Pick which folders should stay matched. New syncs start off.</p><button type="button" onclick="startWizard()">Set up sync</button></div>
</div></div>
{{else}}
<details class="onboarding-disclosure"><summary>Need help setting up another sync? Open the step-by-step guide</summary><div class="quick-start"><div class="quick-start-heading"><h2>New here? Start with these three steps.</h2></div><div class="quick-start-grid">
<div class="quick-step"><span class="step-number">1</span><h3>Connect your cloud files</h3><p>Google Drive or Dropbox, if you use them.</p><button class="secondary" type="button" onclick="startAccountWizard()">Connect account</button></div>
<div class="quick-step"><span class="step-number">2</span><h3>Add your NAS box</h3><p>Give Syncria your NAS address and sign-in details.</p><button class="secondary" type="button" onclick="toggleDirectNAS(true)">Add NAS</button></div>
<div class="quick-step"><span class="step-number">3</span><h3>Choose where files go</h3><p>Choose the folders to match.</p><button class="primary" type="button" onclick="startWizard()">Set up a sync</button></div>
</div></div></details>
{{end}}
<div class="overview-columns grid gap-4">
<section class="overview-list-panel">
 <div class="section-heading"><div><span class="mini-overline">Configured syncs</span><h2>My Syncs</h2><p>Each sync is a connection between places you've chosen.</p></div><button class="text-action" type="button" onclick="navigate('sync')">View all →</button></div>
 <div class="flow-list grid">
 {{range .Jobs}}<div class="flow-row"><span class="flow-row-icon">{{if eq .Provider "bridge"}}☁{{else}}⇄{{end}}</span><div class="flow-row-content"><strong>{{.Name}}</strong><small>{{if eq .Provider "bridge"}}{{if eq .CloudProvider "dropbox"}}Dropbox{{else}}Google Drive{{end}} → NAS{{else if eq .Provider "nas_bridge"}}NAS ↔ NAS{{else}}Existing folder sync{{end}}</small></div>{{if .Enabled}}<span class="flow-state is-on"><span></span> On</span>{{else}}<span class="flow-state is-off"><span></span> Paused</span>{{end}}<button class="flow-row-open" type="button" onclick="openSyncDetails('{{.ID}}')" aria-label="Open {{.Name}} details">↗</button></div>{{else}}<div class="overview-empty"><span>⇄</span><strong>No syncs yet</strong><p>Make your first one to start keeping folders together.</p><button class="secondary" type="button" onclick="startWizard()">Create a sync</button></div>{{end}}
 </div>
</section>
<section class="overview-list-panel">
 <div class="section-heading"><div><span class="mini-overline">Connections</span><h2>Connected storage</h2><p>Cloud accounts and NAS boxes saved in Syncria.</p></div><button class="text-action" type="button" onclick="navigate('storage')">Manage →</button></div>
 <div class="place-summary"><div class="place-summary-icon cloud-symbol">☁</div><div><strong>Cloud accounts</strong><small>Google Drive: {{len .Accounts}} · Dropbox: {{len .DropboxAccounts}}</small></div><button class="flow-row-open" type="button" onclick="navigate('accounts')" aria-label="Open Cloud Accounts">↗</button></div>
 {{range .Shares}}<div class="place-summary"><div class="place-summary-icon">▤</div><div><strong>{{.Label}}</strong><small>{{.Protocol}} · {{.Server}}</small></div>{{if not .Local}}<span class="place-indicator" data-share-mounted="{{.ID}}">NAS connection saved</span>{{else if .Mounted}}<span class="place-indicator" data-share-mounted="{{.ID}}">Mounted</span>{{else}}<span class="place-indicator" data-share-mounted="{{.ID}}">Not mounted</span>{{end}}</div>{{else}}<div class="overview-empty overview-empty-compact"><strong>No NAS boxes connected</strong><p>Add a NAS from Storage to use it in a sync.</p></div>{{end}}
 </section>
</div>
<details class="glossary-fold"><summary>New to syncing? A few words explained.</summary><div class="glossary-layout"><span><strong>NAS</strong> — A NAS is a storage box on your network.</span><span><strong>Cloud</strong> — Online files, like Google Drive.</span><span><strong>Sync</strong> — Keeping folders matched.</span><span><strong>Undo</strong> — Bringing back a recently deleted file.</span></div></details>
</section>
<section class="view" id="view-sync">
<div class="workspace-intro sync-library-header"><div><h2>Connections that keep your files in sync</h2><p>Select a sync to inspect its files and run history, manage its settings, and see what changes next. New syncs start paused.</p></div><button class="primary" type="button" onclick="startWizard()">+ Create sync</button></div>
<div class="sync-toolbar" aria-label="Find and filter syncs"><div class="sync-toolbar-stats">
 <div class="sync-toolbar-stat"><strong id="syncLibraryTotal">{{len .Jobs}}</strong><span>connections</span></div>
 <div class="sync-toolbar-stat"><strong id="syncLibraryRunning">0</strong><span>running</span></div>
 <div class="sync-toolbar-stat"><strong id="syncLibraryAttention">0</strong><span>need attention</span></div>
 </div>
 <div class="sync-toolbar-controls"><label class="sr-only" for="syncLibrarySearch">Find a sync</label><input id="syncLibrarySearch" class="sync-toolbar-search" type="search" placeholder="Search your syncs…" oninput="filterSyncLibrary()">
 <label class="sr-only" for="syncLibraryFilter">Filter by status</label><select id="syncLibraryFilter" onchange="filterSyncLibrary()"><option value="all">All syncs</option><option value="running">Running</option><option value="attention">Needs attention</option><option value="active">Automatic on</option><option value="paused">Paused</option></select></div>
</div>
<p class="sync-library-empty-filter hidden" id="syncLibraryNoMatch">No connections match. Try another search term or status.</p>
<section class="sync-library" aria-label="Your file syncs">
{{range .Jobs}}{{$job := .}}
<article class="relationship-card" data-sync-card="{{.ID}}" data-sync-name="{{.Name}}">
 <div class="flow-card-main">
   <div class="flow-card-emblem">{{if eq .Provider "bridge"}}☁{{else}}⇄{{end}}</div>
   <div class="flow-card-identity"><span class="mini-overline">{{if eq .Provider "bridge"}}CLOUD / NAS{{else if eq .Provider "nas_bridge"}}NAS / NAS{{else}}FOLDER / FOLDER{{end}}</span><h3><button type="button" class="sync-title-link" onclick="openSyncDetails('{{.ID}}')">{{.Name}}</button></h3><p>{{if eq .Provider "bridge"}}{{if eq .CloudProvider "dropbox"}}Dropbox{{else}}Google Drive{{end}} · {{if eq .Direction "one_way"}}One way: cloud → NAS{{else}}Both ways: cloud ↔ NAS{{end}} · {{len .NASDestinations}} NAS destination(s){{else if eq .Provider "nas_bridge"}}{{if eq .Direction "one_way"}}Copy one way{{else}}Keep both ways matched{{end}} · {{len .NASDestinations}} NAS destination(s){{else}}Existing filesystem sync{{end}}</p></div>
   <div class="flow-card-health">{{if .Enabled}}<span class="flow-state is-on" data-job-enabled="{{.ID}}"><span></span> On</span>{{else}}<span class="flow-state is-off" data-job-enabled="{{.ID}}"><span></span> Paused</span>{{end}}</div>
 </div>
 <div class="flow-card-facts">
    <div><small>WHEN IT CHECKS</small><strong>{{if eq .ScheduleMode "changes"}}On NAS changes{{else}}Every {{.Interval}} minutes{{end}}</strong></div>
    <div><small>LAST RUN</small><strong data-job-run="{{.ID}}">{{if .LastRun.IsZero}}Not yet{{else}}{{.LastRun.Format "Jan 2, 2006 · 15:04"}}{{end}}</strong></div>
    <div><small>YOUR FILES</small><strong>{{len .NASDestinations}} NAS folder(s)</strong></div>
 </div>
 <p class="flow-card-description" data-job-summary="{{.ID}}">{{if .LastSummary}}{{.LastSummary}}{{else}}This sync hasn't run yet.{{end}}</p>
 <p class="flow-card-stats" data-job-statline="{{.ID}}">Sent: {{.TotalUploads}} files · Received: {{.TotalDownloads}} files · Successful runs: {{.SuccessfulRuns}} · Errors: {{.FailedRuns}}</p>
 <p class="flow-card-error" data-job-error="{{.ID}}" role="status" {{if not .LastError}}hidden{{end}}>{{if .LastError}}Last sync failed{{if not .LastRun.IsZero}} on {{.LastRun.Format "Jan 2, 2006 15:04"}}{{end}}: {{.LastError}}{{end}}</p>
  <button type="button" class="secondary flow-card-dismiss" data-job-dismiss="{{.ID}}" {{if not .LastError}}hidden{{end}} onclick="dismissSyncFailure('{{.ID}}',this)">Dismiss old failure</button>
 <div class="sync-live-progress hidden" data-job-progress="{{.ID}}">
  <div class="progress-topline"><strong data-job-phase="{{.ID}}">Syncing…</strong><small>Live transfer progress</small></div>
  <progress data-job-bar="{{.ID}}" max="100"></progress><p data-job-progress-detail="{{.ID}}"></p>
 </div>
 {{if .Conflicts}}<section class="relationship-alert" aria-label="Files that need a decision"><div class="alert-title"><span>!</span><div><h4>Two versions need your choice</h4><p>Syncria kept both copies safe until you choose.</p></div></div>
 {{range .Conflicts}}<div class="decision-row"><div><strong>{{.Path}}</strong><small>First NAS: {{.PrimaryModified.Format "Jan 2, 2006 · 15:04"}} · Other NAS: {{.SecondaryModified.Format "Jan 2, 2006 · 15:04"}}</small></div>
 <form method="post" action="/api/conflicts/resolve" class="inline"><input type="hidden" name="id" value="{{$job.ID}}"><input type="hidden" name="pair" value="{{.Pair}}"><input type="hidden" name="path" value="{{.Path}}">
 <label class="field"><span class="sr-only">Which copy to keep</span><select name="choice"><option value="primary">Use first NAS copy</option><option value="secondary">Use other NAS copy</option><option value="keep_both">Keep both copies</option></select></label><button type="submit" class="secondary">Save choice</button></form></div>{{end}}
 </section>{{end}}
 {{if .PendingDeletions}}
 <div class="recovery-summary" data-recovery-summary="{{.ID}}">
  <div class="recovery-summary-main"><span class="recovery-symbol" aria-hidden="true">↶</span><div><strong>Recently deleted <span class="recovery-count">{{len .PendingDeletions}}</span></strong><small>Files can be restored during their recovery window.</small></div></div>
  <button type="button" class="secondary" onclick="openRecoveryDialog('{{.ID}}')">Review &amp; restore <span aria-hidden="true">→</span></button>
 </div>
 <dialog class="recovery-dialog" id="recovery-{{.ID}}" aria-labelledby="recovery-title-{{.ID}}" aria-describedby="recovery-help-{{.ID}}">
  <div class="recovery-dialog-shell">
   <header class="recovery-dialog-header"><div><span class="mini-overline">SYNC FILE RECOVERY</span><h2 id="recovery-title-{{.ID}}">Recently deleted files</h2><p id="recovery-help-{{.ID}}">Restore individual files while their recovery window is still open. No other files are changed.</p></div>
    <button type="button" class="secondary recovery-close" onclick="closeRecoveryDialog('{{.ID}}')" aria-label="Close recently deleted files">Close <span aria-hidden="true">×</span></button>
   </header>
   <div class="recovery-dialog-toolbar"><label for="recovery-search-{{.ID}}">Find a file <input id="recovery-search-{{.ID}}" type="search" autocomplete="off" placeholder="Search filename or folder…" oninput="filterRecoveryFiles('{{.ID}}')"></label>
    <span class="recovery-visible-count" id="recovery-visible-{{.ID}}" role="status">{{len .PendingDeletions}} files eligible for recovery</span>
   </div>
   <div class="recovery-dialog-items" id="recovery-items-{{.ID}}">
    {{range .PendingDeletions}}<div class="recovery-dialog-item" data-recovery-path="{{.Path}}" data-recovery-expires-ms="{{.Expires.UnixMilli}}"><div class="recovery-dialog-item-copy"><strong>{{.Path}}</strong><small>Undo before {{.Expires.Format "Jan 2, 2006 · 15:04:05 MST"}}</small></div>
     <form method="post" action="/api/recovery/undo" class="inline"><input type="hidden" name="operation" value="undo"><input type="hidden" name="id" value="{{$job.ID}}"><input type="hidden" name="pair" value="{{.Pair}}"><input type="hidden" name="path" value="{{.Path}}"><button type="submit" class="secondary">Restore file</button></form></div>{{end}}
   </div>
   <div class="recovery-dialog-empty hidden" id="recovery-empty-{{.ID}}">No filenames match. Try another search.</div>
   <footer class="recovery-dialog-footer"><span>Recovery is temporary; expired files may no longer be restorable.</span><button type="button" class="secondary" onclick="closeRecoveryDialog('{{.ID}}')">Done</button></footer>
  </div>
 </dialog>
 {{end}}
 <section class="sync-preview hidden" data-preview="{{.ID}}" aria-label="Read-only sync preview">
  <div class="sync-preview-heading"><div><strong>Test this sync without changing files</strong><p>See which files would move, which side wins, and what needs a decision.</p></div><button class="secondary" type="button" onclick="previewSync('{{.ID}}')">Refresh preview</button></div>
  <div class="sync-preview-result" data-preview-result="{{.ID}}" role="status">Scanning both locations…</div>
 </section>
 <details class="sync-run-events hidden" data-job-events="{{.ID}}"><summary>What actually changed in the last run</summary><div class="sync-run-event-list" data-job-events-list="{{.ID}}"></div></details>
 <div class="flow-card-actions">
  <button type="button" class="secondary sync-action-open" onclick="openSyncDetails('{{.ID}}')">Open sync <span aria-hidden="true">→</span></button>
  <form method="post" action="/action" class="inline"><input type="hidden" name="action" value="run_job"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="primary" data-job-run-control="{{.ID}}">Sync now</button></form>
  <button type="button" class="secondary hidden" data-job-cancel="{{.ID}}" onclick="requestStopSync('{{.ID}}',this)">Stop run</button>
  <span class="sync-action-spacer"></span>
  <details class="sync-card-more" data-sync-menu="{{.ID}}"><summary aria-label="More actions for {{.Name}}">More <span aria-hidden="true">⋯</span></summary>
   <div class="sync-card-more-menu" role="group" aria-label="More sync actions">
    <button type="button" onclick="openSyncDetails('{{.ID}}','files')">Preview file changes</button>
    {{if or (eq .Provider "bridge") (eq .Provider "nas_bridge")}}<button type="button" onclick="editSync('{{.ID}}')">Edit sync configuration</button>{{end}}
    <form method="post" action="/action"><input type="hidden" name="action" value="toggle_job"><input type="hidden" name="id" value="{{.ID}}"><button type="submit">{{if .Enabled}}Pause automatic checks{{else}}Enable automatic checks{{end}}</button></form>
    <form method="post" action="/action" onsubmit="return confirm('Remove this sync configuration? Your original files on cloud and NAS are not deleted.')"><input type="hidden" name="action" value="delete_job"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="danger">Remove connection</button></form>
   </div>
  </details>
 </div>
</article>
{{else}}
<div class="flow-library-empty"><span class="empty-symbol" aria-hidden="true">⇄</span><h3>No syncs configured</h3><p>Connect two folders, then decide whether and when Syncria should keep them matched.</p><button type="button" class="primary" onclick="startWizard()">Create my first sync →</button></div>
{{end}}
</section>
<details class="glossary-fold"><summary>Questions about one-way, both ways, or Undo?</summary><div class="glossary-layout"><span><strong>One way</strong> copies changes from the first NAS to the others.</span><span><strong>Both ways</strong> lets changes move between NAS boxes.</span><span><strong>Undo</strong> can bring back a recent deletion.</span></div></details>
<section id="wizard" class="sync-builder hidden"><div class="builder-heading"><div><span class="mini-overline" id="wizardModeLabel">New sync</span><h2 id="wizardHeading">Connect your files.</h2><p>Four short steps. No files move until you're ready.</p></div><button type="button" class="secondary" onclick="closeSyncWizard()">Close ✕</button></div>
<div class="step-indicator"><span id="stepMarker1" class="active">1 · Files start here</span><span id="stepMarker2">2 · Copy them here</span><span id="stepMarker3">3 · Pick the rules</span><span id="stepMarker4">4 · Double-check</span></div>
<form id="syncWizard" method="post" action="/action" onsubmit="return submitSyncWizard(event)" novalidate><input type="hidden" name="action" value="add_bridge"><input type="hidden" id="editingSyncId" name="edit_id" value=""><input type="hidden" id="editingSyncType" name="relationship_type" value="" disabled>
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
<div id="cloudDirectionFields" class="panel"><label class="field">Which direction should files sync?
<span class="field-help">One way copies from the cloud to your NAS. NAS edits will not be uploaded, and differing NAS files may be replaced by cloud copies. Both ways lets changes flow in either direction.</span>
<select id="cloudDirection" name="cloud_direction" onchange="cloudDirectionChanged()">
<option value="bidirectional">Both ways: cloud ↔ NAS</option>
<option value="one_way">One way: cloud → NAS only</option>
</select></label><p class="muted">Review the dry run before the first live sync. One-way means the cloud is authoritative; it does not mean that existing NAS files will never be changed.</p></div>
<div id="cloudFirstRunFields" class="panel"><label class="field">If the same file is already on both sides when this sync starts
<span class="field-help">Syncria compares actual file contents before replacing anything. When the contents differ, decide how to choose the first copy.</span>
<select id="initialConflictPolicy" name="initial_conflict_policy">
<option value="newer">Use the file with the newer modification time (recommended)</option>
<option value="manual">Ask me which copy to keep</option>
</select></label>
<p class="muted">Check the dry run first. Copying files between devices can change their modification times, so newer does not always mean better. Later conflicts still require their own resolution.</p></div>
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
<p class="wizard-status good" id="reviewStatus">Your sync starts paused. No files move when you save these settings.</p>
<div id="editMappingWarning" class="panel hidden"><h3>Changing connected folders?</h3><p class="muted">Changing the source, NAS destination, or folder requires Syncria to establish a fresh comparison baseline. Old sync checkpoints cannot safely be applied to different folders. This does not delete files; preview the new mapping before syncing.</p>
<label><input id="editResetMapping" type="checkbox" name="reset_mapping" value="1"> I understand: reset the file comparison baseline if I changed a source or destination</label></div>
<div class="row"><button type="button" class="secondary" onclick="showSyncStep(3)">← Back</button><button type="submit" class="primary" id="createSyncButton">Save my sync (keep it off)</button></div></div></form></section></section>
<section class="view" id="view-storage">
<div class="workspace-intro"><div><p>Connect and manage the network storage used by your syncs. Adding a NAS does not transfer files.</p></div><button class="primary" type="button" onclick="toggleDirectNAS()">+ Add NAS</button></div>
<section id="directNASPanel" class="connection-drawer hidden" aria-label="Add a NAS connection"><div class="drawer-heading"><div><span class="mini-overline">New NAS</span><h3>Connect a NAS box</h3><p>You'll need the address, shared folder, and sign-in details from your NAS.</p></div><button type="button" class="secondary" onclick="toggleDirectNAS(false)">Close ✕</button></div>
<form method="post" action="/api/nas/action"><input type="hidden" name="operation" value="connect">
<div class="settings-grid">
<label class="field">Give it a name<span class="field-help">Something you'll recognize, like “Office NAS”.</span><input name="label" id="shareLabel" placeholder="Family NAS" maxlength="128" required></label>
<label class="field">How does it share files?<span class="field-help">The default works for most NAS boxes.</span><select name="protocol" id="shareProtocol"><option value="smb">Windows-style shared folders (SMB) — recommended</option><option value="nfs" disabled>Other network sharing (NFS) — not available yet</option></select></label>
<label class="field">Network address<span class="field-help">Usually looks like 192.168.1.20.</span><input name="server" id="shareServer" placeholder="192.168.30.10" required></label>
<label class="field">Shared folder<span class="field-help">The folder name your NAS makes available.</span><input name="export" id="sharePath" placeholder="Documents" required></label>
<label class="field">NAS username<input name="username" required autocomplete="username"></label>
<label class="field">NAS password<input type="password" name="password" required autocomplete="new-password"></label>
</div><div class="drawer-footer"><p>Syncria checks the connection before saving it.</p><button class="primary" type="submit">Test & save connection <span aria-hidden="true">→</span></button></div></form></section>
<section class="connection-list"><div class="section-heading"><div><span class="mini-overline">Saved NAS devices</span><h2>My NAS Boxes</h2><p>Each saved connection can be used by one or more syncs.</p></div><span class="section-counter">{{len .Shares}} saved</span></div>
{{range .Shares}}<div class="connection-row"><div class="connection-emblem"><svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.7"><rect x="3" y="3" width="18" height="8" rx="2"/><rect x="3" y="13" width="18" height="8" rx="2"/><path d="M7 7h.01M7 17h.01"/></svg></div><div class="connection-copy"><strong>{{.Label}}</strong><p>{{.Server}} / {{.Export}}</p><small>{{if .Local}}Older mounted connection — kept as configured{{else}}Direct SMB connection{{end}}</small></div><div class="connection-tools">
{{if not .Local}}<span class="connection-state" data-share-mounted="{{.ID}}">Saved connection</span>
<form class="inline" action="/api/nas/action" method="post"><input type="hidden" name="operation" value="verify"><input type="hidden" name="id" value="{{.ID}}"><button class="secondary" type="submit">Test</button></form>
<form class="inline" action="/api/nas/action" method="post" onsubmit="return confirm('Remove this NAS connection? Files on the NAS will not be touched.')"><input type="hidden" name="operation" value="disconnect"><input type="hidden" name="id" value="{{.ID}}"><button class="danger" type="submit">Remove</button></form>{{else}}<span class="connection-state" data-share-mounted="{{.ID}}">{{if .Mounted}}Mounted{{else}}Not mounted{{end}}</span>{{end}}</div></div>
{{else}}<div class="flow-library-empty"><span class="empty-symbol">▤</span><span class="mini-overline">No NAS devices</span><h3>Your NAS boxes will appear here.</h3><p>Add one to get started. Files won't be copied until you set up a sync.</p><button type="button" class="primary" onclick="toggleDirectNAS(true)">Add a NAS →</button></div>{{end}}</section>
</section>
<section class="view" id="view-accounts">
<div class="workspace-intro"><div><p>Connect Google Drive or Dropbox accounts. Each account stays separate until you choose folders to sync.</p></div><button class="primary" type="button" onclick="startAccountWizard()">+ Connect account</button></div>
<div class="provider-grid grid gap-4">
<section class="provider-panel"><div class="provider-top"><span class="provider-icon google-icon">G</span><div><span class="mini-overline">GOOGLE</span><h3>Google Drive</h3><p>Use files from one or more Google accounts.</p></div><span class="provider-count">{{len .Accounts}} saved</span></div>
<div class="provider-list">{{range .Accounts}}<div class="provider-account"><div><strong>{{.Email}}</strong><small>Connected to Syncria</small></div><div class="provider-buttons"><button class="secondary" type="button" onclick="verifyConnectedAccount('{{.ID}}',this)">Test</button><form method="post" action="/action" class="inline" onsubmit="return confirm('Disconnect this Google Drive account? Your files will stay on Google Drive.')"><input type="hidden" name="action" value="remove_account"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="danger">Remove</button></form></div></div>{{else}}<div class="provider-empty">No Google Drive accounts connected yet.</div>{{end}}</div>
<button type="button" class="provider-add" onclick="startAccountWizard('google_drive')">+ Add Google Drive account <span aria-hidden="true">↗</span></button></section>
<section class="provider-panel"><div class="provider-top"><span class="provider-icon dropbox-icon">◆</span><div><span class="mini-overline">DROPBOX</span><h3>Dropbox</h3><p>Bring your Dropbox folders into a sync.</p></div><span class="provider-count">{{len .DropboxAccounts}} saved</span></div>
<div class="provider-list">{{range .DropboxAccounts}}<div class="provider-account"><div><strong>{{.Email}}</strong><small>Connected to Syncria</small></div><div class="provider-buttons"><form method="post" action="/action" class="inline" onsubmit="return confirm('Disconnect Dropbox? Your files stay in Dropbox.')"><input type="hidden" name="action" value="remove_dropbox_account"><input type="hidden" name="id" value="{{.ID}}"><button type="submit" class="danger">Remove</button></form></div></div>{{else}}<div class="provider-empty">No Dropbox accounts connected yet.</div>{{end}}</div>
<button type="button" class="provider-add" onclick="startAccountWizard('dropbox')">+ Add Dropbox account <span aria-hidden="true">↗</span></button></section>
</div>
<section class="connection-drawer hidden" id="accountWizard"><div class="drawer-heading"><div><span class="mini-overline">Connect an account</span><h3>Connect your cloud files</h3><p>Choose a service and sign in securely to connect it.</p></div><button class="secondary" type="button" onclick="$('accountWizard').classList.add('hidden')">Close ✕</button></div>
<div class="step-indicator"><span id="accountMarker1" class="active">1 · Choose service</span><span id="accountMarker2">2 · Sign in</span></div>
<div class="step active" id="accountStep1"><label class="field">Where are your cloud files?<select id="accountProvider" onchange="accountProviderChanged()"><option value="google_drive">Google Drive</option><option value="dropbox">Dropbox</option></select></label><p class="muted">This saves a connection, not your files. You'll choose folders when you create a sync.</p><div class="row"><button type="button" class="primary" onclick="showAccountStep(2)">Continue →</button><button type="button" class="secondary" onclick="navigate('settings')">Advanced sign-in settings</button></div></div>
<div class="step" id="accountStep2"><div id="googleAccountAuthorize"><p>Google Drive needs sign-in settings first. The person running Syncria can set them up in Settings.</p>
{{if .ClientID}}<a class="primary" href="/oauth/start">Sign in with Google Drive →</a>{{else}}<p class="wizard-status bad">Google sign-in is not set up yet.</p><button type="button" class="secondary" onclick="navigate('settings')">Open sign-in settings</button>{{end}}</div>
<div class="hidden" id="dropboxAccountAuthorize"><p>Dropbox needs sign-in settings first. The person running Syncria can set them up in Settings.</p>{{if .DropboxClientID}}<a class="primary" href="/oauth/dropbox/start">Sign in with Dropbox →</a>{{else}}<p class="wizard-status bad">Dropbox sign-in is not set up yet.</p><button type="button" class="secondary" onclick="navigate('settings')">Open sign-in settings</button>{{end}}</div><div class="row"><button type="button" class="secondary" onclick="showAccountStep(1)">← Back</button></div></div>
</section>
</section>
<section class="view" id="view-sync-detail">
 <div class="sync-detail-back"><button type="button" class="secondary" onclick="navigate('sync')">← All syncs</button><span class="sync-library-hint">Details update automatically as synchronization runs.</span></div>
 <section class="panel sync-details-shell" aria-label="Selected synchronization">
  <div class="sync-details-head">
   <div><span class="mini-overline">SYNCHRONIZATION</span><h2 class="sync-details-name" id="syncDetailTitle">Sync details</h2><p class="sync-details-meta" id="syncDetailSubtitle">Choose a sync to inspect its activity and settings.</p></div>
   <div class="sync-status-row"><span class="sync-status-pill" id="syncDetailStatus" data-kind="healthy">Loading</span></div>
  </div>
  <div class="sync-details-toolbar">
   <button type="button" class="primary" id="syncDetailRun" onclick="submitSyncAction('run_job')">Sync now</button>
   <button type="button" class="secondary hidden" id="syncDetailStop" onclick="requestStopSync(selectedSyncID,this)">Stop current run</button>
   <button type="button" class="secondary" id="syncDetailPause" onclick="submitSyncAction('toggle_job')">Pause automatic checks</button>
   <button type="button" class="secondary" id="syncDetailEdit" onclick="editSync(selectedSyncID)">Edit configuration</button>
  </div>
  <div class="sync-details-tabs" role="tablist" aria-label="Sync detail sections">
   <button type="button" role="tab" aria-controls="syncDetailOverview" id="syncTabOverview" aria-selected="true" onclick="showSyncDetailTab('overview')">Overview</button>
   <button type="button" role="tab" aria-controls="syncDetailFiles" id="syncTabFiles" aria-selected="false" onclick="showSyncDetailTab('files')">Files &amp; preview</button>
   <button type="button" role="tab" aria-controls="syncDetailHistory" id="syncTabHistory" aria-selected="false" onclick="showSyncDetailTab('history')">Run history</button>
   <button type="button" role="tab" aria-controls="syncDetailSettings" id="syncTabSettings" aria-selected="false" onclick="showSyncDetailTab('settings')">Settings</button>
  </div>
  <section class="sync-details-tab is-active" id="syncDetailOverview" role="tabpanel" aria-labelledby="syncTabOverview">
   <div class="sync-detail-grid">
    <div class="sync-detail-stat"><span>Last successful checks</span><strong id="syncDetailSuccess">—</strong></div>
    <div class="sync-detail-stat"><span>Files moved overall</span><strong id="syncDetailMoved">—</strong></div>
    <div class="sync-detail-stat"><span>Data transferred</span><strong id="syncDetailBytes">—</strong></div>
   </div>
   <div class="sync-detail-section"><h3>Current status</h3><div id="syncDetailLive" class="sync-detail-live" aria-live="polite">Loading live status…</div><p class="sync-detail-note" id="syncDetailWatchNote">Synchronization checks use saved file metadata and change detection where supported.</p></div>
   <div class="sync-detail-section"><div class="sync-run-history-heading"><h3>Latest run</h3><button type="button" class="secondary" onclick="showSyncDetailTab('history')">Browse every run →</button></div><div id="syncDetailLastRun" class="sync-detail-empty"><strong>No detailed run loaded yet</strong><p>The most recent result will appear here. Older versions recorded totals rather than full file-by-file history.</p></div></div>
   <div class="sync-detail-section hidden" id="syncDetailRecoveryBar"><div class="sync-run-history-heading"><h3>Recently deleted files</h3><button type="button" class="secondary" onclick="openRecoveryDialog(selectedSyncID)">Review &amp; restore <span id="syncDetailRecoveryCount"></span> →</button></div><p class="sync-detail-note">Recover eligible files before their undo window expires.</p></div>
   <div class="sync-detail-section"><h3>What happens to your files</h3><div class="sync-detail-pairs">
    <div class="sync-detail-pair"><small>Transfer direction</small><strong id="syncDetailDirection">—</strong></div>
    <div class="sync-detail-pair"><small>Deletion behavior</small><strong id="syncDetailDeletion">—</strong></div>
   </div></div>
  </section>
  <section class="sync-details-tab" id="syncDetailFiles" role="tabpanel" aria-labelledby="syncTabFiles">
   <div class="sync-preview-head"><div><h3>Planned file changes</h3><p class="muted">Compare file metadata without copying anything. See replacements, skips, deletions and explanations—not unchanged files.</p></div><button type="button" class="secondary" onclick="runDetailPreview()">Run read-only preview</button></div>
   <div id="syncDetailPreviewHost" class="sync-detail-preview-host"></div>
   <p class="sync-detail-note">Looking for files that already transferred? Open Run history and select a completed run.</p>
  </section>
  <section class="sync-details-tab" id="syncDetailHistory" role="tabpanel" aria-labelledby="syncTabHistory">
   <div class="sync-run-history-heading"><div><h3>Recorded sync runs</h3><p class="muted">Select any run to see the actual files and operations recorded for it.</p></div><button type="button" class="secondary" onclick="loadSyncRunList()">Refresh history</button></div>
   <div class="sync-run-history-layout">
    <div class="sync-detail-run-column"><div id="syncRunList" class="sync-run-list" aria-label="Run history"></div><div id="syncRunPager" class="sync-detail-pager"></div></div>
    <div class="sync-detail-files-column">
     <div class="sync-run-history-heading"><h3>Files in selected run</h3></div>
     <p id="syncRunSummary" class="muted">Choose a run to inspect recorded changes.</p>
     <div class="sync-detail-files-toolbar"><div class="sync-file-filter">
      <label class="sr-only" for="syncFileSearch">Search filenames in the selected run</label><input id="syncFileSearch" type="search" placeholder="Find filename or folder…" oninput="debounceRunFileSearch()">
      <label class="sr-only" for="syncFileActionFilter">Filter file operations</label><select id="syncFileActionFilter" onchange="reloadRunFileFilters()"><option value="">All operations</option><option value="copy">Copied</option><option value="replace">Replaced</option><option value="delete">Deleted</option><option value="transfer failed">Failed</option><option value="unsupported Google document — kept on Drive">Skipped unsupported</option></select>
     </div></div>
     <div id="syncRunEvents" class="sync-run-files" aria-label="Files in selected run"></div><div id="syncEventPager" class="sync-detail-pager"></div>
    </div>
   </div>
  </section>
  <section class="sync-details-tab" id="syncDetailSettings" role="tabpanel" aria-labelledby="syncTabSettings">
   <h3>Relationship settings</h3><p class="muted">What Syncria does, when it runs, and how it handles deletions.</p>
   <div class="sync-detail-pairs">
    <div class="sync-detail-pair"><small>Sync direction</small><strong id="syncSettingDirection">—</strong></div>
    <div class="sync-detail-pair"><small>Checking schedule</small><strong id="syncSettingSchedule">—</strong></div>
    <div class="sync-detail-pair"><small>Deletion sharing</small><strong id="syncSettingDelete">—</strong></div>
    <div class="sync-detail-pair"><small>Undo protection</small><strong id="syncSettingRecovery">—</strong></div>
   </div>
   <p class="sync-detail-alert" id="syncDeleteWarning" role="note">Review how deletions are shared before enabling this sync.</p>
   <div class="sync-detail-section"><h3>Manage this connection</h3><div class="sync-settings-actions">
    <button type="button" class="primary" onclick="editSync(selectedSyncID)">Edit source, destination and policies</button>
    <button type="button" class="secondary" onclick="showSyncDetailTab('files')">Preview before syncing</button>
    <button type="button" class="danger" onclick="submitSyncAction('delete_job')">Remove sync configuration</button>
   </div><p class="sync-detail-note">Removing a sync configuration does not delete the underlying cloud or NAS files. Changes to deletion settings only take effect when saved.</p></div>
  </section>
 </section>
</section>
<section class="view" id="view-activity">
<section class="panel live-sync-workspace" aria-label="Active synchronization runs"><div class="section-heading"><div><span class="mini-overline">WORK IN PROGRESS</span><h2>Active syncs</h2><p>Watch what needs copying, what is transferring, and stop a run without restarting Syncria.</p></div><span class="pill">Live</span></div><div id="activeSyncList" class="active-sync-list"><p class="muted">No sync is running right now.</p></div></section>
<section class="panel live-activity-panel"><div class="section-heading"><div><h2>Live file activity</h2><p>Changes appear automatically as files are copied, replaced, or removed.</p></div><span class="pill">Live</span></div><div id="liveActivityFeed" class="live-activity-feed" aria-live="off"><p class="muted">Waiting for activity…</p></div></section>
<div class="workspace-intro"><div><p>Transfer counts and recent results, updated as your syncs run.</p></div><button class="secondary" type="button" onclick="refreshDashboardStatus()">Refresh activity ↻</button></div>
<div class="activity-scoreboard grid gap-3">
 <div class="activity-score activity-score-highlight"><small>CHECKS THAT WORKED</small><strong id="activitySuccessful">{{.Metrics.Runs}}</strong><span>completed checks</span></div>
 <div class="activity-score"><small>CHECKS WITH ERRORS</small><strong id="activityFailed">{{.Metrics.Failures}}</strong><span>historical failed checks</span></div>
 <div class="activity-score"><small>FILES COPIED</small><strong id="activityFiles">—</strong><span>across all syncs</span></div>
 <div class="activity-score"><small>FILES SENT</small><strong id="activityUploaded">—</strong><span>to another place</span></div>
 <div class="activity-score"><small>FILES RECEIVED</small><strong id="activityDownloaded">—</strong><span>from another place</span></div>
 <div class="activity-score"><small>DATA TRANSFERRED</small><strong id="activityTransfer">—</strong><span>total</span></div>
</div>
<section class="activity-table-shell"><div class="section-heading"><div><span class="mini-overline">Sync results</span><h2>Sync details</h2><p>Numbers come from Syncria's actual transfer records.</p></div></div>
{{range .Jobs}}<div class="activity-row"><div class="activity-row-head"><div><strong><button type="button" class="sync-title-link" onclick="openSyncDetails('{{.ID}}')">{{.Name}}</button></strong><small>{{if eq .Provider "bridge"}}{{if eq .CloudProvider "dropbox"}}Dropbox{{else}}Google Drive{{end}} ↔ NAS{{else if eq .Provider "nas_bridge"}}NAS ↔ NAS{{else}}Existing folder sync{{end}}</small></div><span class="activity-state" data-activity-running="{{.ID}}">Loading results…</span></div>
<div class="activity-row-numbers">
 <div><small>GOOD CHECKS</small><strong data-job-runs="{{.ID}}">{{.SuccessfulRuns}}</strong></div>
 <div><small>ERRORS</small><strong data-job-fails="{{.ID}}">{{.FailedRuns}}</strong></div>
 <div><small>FILES MOVED</small><strong data-job-files="{{.ID}}">—</strong></div>
 <div><small>BYTES MOVED</small><strong data-job-bytes="{{.ID}}">—</strong></div>
</div><p class="activity-row-detail" data-job-detail="{{.ID}}">Loading most recent result…</p><button type="button" class="secondary" onclick="openSyncDetails('{{.ID}}')">View individual run history</button></div>{{else}}<div class="flow-library-empty"><strong>No activity yet</strong><p>Create a sync to see its results here.</p></div>{{end}}</section>
</section>
<section class="view" id="view-updates">
<div class="workspace-intro"><div><p>Install updates here. Download, verification, and restart progress appear when an update runs.</p></div></div>
<div class="update-console">
<section class="panel update-primary" id="updateStatusCard" data-update-state="checking" aria-live="polite">
 <div class="update-status-header"><span class="update-status-icon" id="updateStatusIcon" aria-hidden="true">↻</span><div class="update-status-text"><span class="mini-overline">Application status</span><h2 id="updateStatusHeadline">Checking for updates…</h2><p id="updateMessage">Checking the latest verified release for this device.</p></div><button class="secondary" type="button" onclick="checkUpdate(true)">Check for updates</button></div>
 <div class="update-version-line"><div><small>INSTALLED VERSION</small><strong id="installedVersion">{{.Version}}</strong></div><span class="update-version-arrow" aria-hidden="true">→</span><div><small>LATEST RELEASE</small><strong id="updateAvailableVersion">Checking…</strong></div></div>
 <p class="update-release-link"><a id="updateReleaseNotes" href="https://github.com/gigabytegrove/syncria/releases" target="_blank" rel="noopener noreferrer">See release notes ↗</a></p>
 <div class="release-actions"><button id="installUpdateButton" class="primary" type="button" onclick="startUpdate('install')" disabled>Install update</button></div>
 <span class="sr-only" id="latestVersion">Checking GitHub Releases…</span>
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
</div></section>

<section class="panel hidden" id="updateHistoryPanel"><div class="panel-head"><h2>Live update log</h2><span class="status" id="updateLogState">Waiting</span></div><p class="muted">Real events recorded by Syncria, with timestamps.</p><ol class="update-timeline" id="updateEventTimeline" aria-live="polite"><li><strong>Ready</strong><small>Events appear here during an update.</small></li></ol></section>
<section class="update-advanced-section"><details class="advanced-update"><summary>Advanced options</summary><div class="advanced-update-body"><div class="update-rollback-section"><h2>Restore previous version</h2><p class="muted">Return to the saved executable if a recent update has caused problems. Your sync settings are retained.</p><button id="rollbackUpdateButton" class="secondary" type="button" onclick="startUpdate('rollback')">Restore previous version</button></div><p class="muted">For troubleshooting or installations without access to automatic updates. Most people won't need these options.</p><h2>Manual installation</h2><p class="muted">Use this only when automatic downloads are unavailable. Supply a matching app file and SHA256SUMS. Your previous version remains available for rollback.</p><form id="manualUpdateForm" action="/api/update/manual" method="post" enctype="multipart/form-data" onsubmit="return submitManualUpdate(event)"><label class="field">App executable<input type="file" name="binary" id="manualBinary" required></label><label class="field">SHA256SUMS<input type="file" name="checksums" required></label><button id="manualUpdateButton" class="secondary" type="submit">Verify & install files</button></form></div></details></section>
</div>
</section>

<section class="view" id="view-computers">
 <div class="workspace-intro"><div><span class="mini-overline">DIRECT CONNECTIONS</span><h2>Computers on your network</h2><p>Browse and transfer files between Windows and Linux computers without SMB or NFS. The files travel directly between the computers, not through Syncria's server.</p></div><button type="button" class="secondary" onclick="loadComputers()">Refresh computers</button></div>
 <div class="computer-layout">
  <section class="panel"><div class="section-heading"><div><h2>Connected computers</h2><p>Syncria verifies each agent's pairing key and TLS certificate before connecting.</p></div></div><div id="computerDeviceList" class="computer-device-list"><p class="muted">Open this page to check connected computers.</p></div></section>
  <section class="panel"><div class="section-heading"><div><h2>Add a computer</h2><p>Run the Syncria executable on the computer you're sharing. Use a dedicated folder and agent identity directory.</p></div></div>
   <div class="computer-help"><strong>Windows (PowerShell)</strong><code>.\syncria-windows-amd64.exe -agent -listen 0.0.0.0:9765 -agent-data "C:\SyncriaAgent" -agent-name "My PC" -agent-root "Files=C:\Users\Public\Documents"</code><strong>Linux</strong><code>./syncria-linux-amd64 -agent -listen 0.0.0.0:9765 -agent-data "$HOME/.config/syncria-agent" -agent-name "My PC" -agent-root "Files=$HOME/Documents"</code><small>Allow inbound TCP 9765 from your trusted LAN. The agent displays its pairing key and TLS SHA-256 fingerprint on first start. Keep both private. For additional folders, repeat <code>-agent-root Name=/path</code>. Running this as a normal user limits access to that user's files.</small></div>
   <form id="computerPairForm" onsubmit="event.preventDefault();pairComputer()" class="computer-pair-form">
    <label class="field">Computer name<input id="computerPairLabel" autocomplete="off" placeholder="Living room desktop" required maxlength="80"></label>
    <label class="field">HTTPS address<input id="computerPairURL" autocomplete="off" placeholder="https://192.168.1.42:9765" required></label>
    <label class="field">Pairing key<input id="computerPairKey" type="password" autocomplete="off" required></label>
    <label class="field">TLS SHA-256 fingerprint<input id="computerPairFingerprint" autocomplete="off" placeholder="64 hexadecimal characters" required></label>
    <button id="computerPairButton" type="submit" class="primary">Connect computer</button><p id="computerPairMessage" class="muted" role="status"></p>
   </form>
  </section>
 </div>
 <section class="panel computer-transfer-workspace"><div class="section-heading"><div><h2>Transfer between computers</h2><p>Choose a file on the left and a folder on the right. A move deletes the original only after the destination verifies the complete file.</p></div></div>
  <div class="computer-explorer">
   <div class="computer-explorer-pane"><h3>From this computer</h3><label class="field">Computer<select id="computerSourceID" onchange="computerSideChanged('source')"><option value="">Select source</option></select></label><label class="field">Shared folder<select id="computerSourceRoot" onchange="computerRootChanged('source')"><option value="">Choose folder</option></select></label><div class="computer-path"><button type="button" class="secondary" onclick="computerUp('source')">↑ Up</button><span id="computerSourcePath">Choose a computer</span></div><div class="computer-file-list" id="computerSourceFiles" role="list"><p class="muted">Select a source computer.</p></div></div>
   <div class="computer-explorer-pane"><h3>To this computer</h3><label class="field">Computer<select id="computerDestinationID" onchange="computerSideChanged('destination')"><option value="">Select destination</option></select></label><label class="field">Shared folder<select id="computerDestinationRoot" onchange="computerRootChanged('destination')"><option value="">Choose folder</option></select></label><div class="computer-path"><button type="button" class="secondary" onclick="computerUp('destination')">↑ Up</button><span id="computerDestinationPath">Choose a computer</span></div><div class="computer-file-list" id="computerDestinationFiles" role="list"><p class="muted">Select a destination computer.</p></div></div>
  </div><div class="computer-transfer-actions"><div><strong id="computerSelectedFile">No file selected</strong><small id="computerTransferMessage">Destination files are never silently overwritten.</small></div><div class="row"><button type="button" class="secondary" onclick="startComputerTransfer(false)">Copy file →</button><button type="button" class="primary" onclick="startComputerTransfer(true)">Move file →</button></div></div>
 </section>
 <section class="panel"><div class="section-heading"><div><h2>Recent direct transfers</h2><p>Actual copy/move operations and their results; a failed move never silently removes the source.</p></div><button type="button" class="secondary" onclick="refreshComputerTransfers()">Refresh transfers</button></div><div id="computerTransferList" class="computer-transfer-list"><p class="muted">No transfers started in this session.</p></div></section>
</section>
<section class="view" id="view-settings"><div class="workspace-intro"><div><p>Manage account credentials and connection settings. These usually only need to be configured once.</p></div></div><section class="panel"><h2>Google sign-in settings (advanced)</h2><p>For an administrator: create a Google Cloud web sign-in app, enable the Google Drive API, and add the address shown below as its redirect URL:</p><div class="selected"><code>{{.Callback}}</code></div><form method="post" action="/action"><input type="hidden" name="action" value="oauth_settings"><label class="field">Google app ID (Client ID)<input name="client_id" value="{{.ClientID}}" required></label><label class="field">Google app secret<input name="client_secret" type="password" placeholder="Leave blank to keep existing secret"></label><label class="field">Your Syncria web address<input name="base_url" value="{{.BaseURL}}" required></label><button class="primary">Save Google sign-in settings</button></form></section><section class="panel"><h2>Dropbox sign-in settings (advanced)</h2>
<p>For an administrator: add this return address to your Dropbox developer app:</p><div class="selected"><code>{{.DropboxCallback}}</code></div>
<form method="post" action="/action"><input type="hidden" name="action" value="dropbox_oauth_settings">
<label class="field">Dropbox app key<input name="client_id" value="{{.DropboxClientID}}" required></label>
<label class="field">Dropbox app secret<input name="client_secret" type="password" placeholder="Leave empty to keep existing secret"></label>
<button type="submit" class="primary">Save Dropbox sign-in settings</button></form></section></section>
{{end}}</div></main>
{{if and .Configured .Logged}}<div class="picker" id="picker" role="dialog" aria-modal="true" aria-labelledby="pickerTitle"><div class="picker-box"><div class="picker-head"><div class="panel-head" style="margin:0"><h2 id="pickerTitle">Select folder</h2><button class="secondary" type="button" onclick="closePicker()">✕ Close</button></div><div id="pickerPath" class="picker-path">Loading…</div></div><div class="row picker-actions"><button class="secondary" type="button" id="pickerUp" onclick="pickerUp()">← Back</button><button class="secondary" type="button" id="pickerNew" onclick="createFolder()">+ New folder</button></div><div id="folderList" class="folder-list" aria-live="polite"></div><div class="picker-footer"><small id="pickerStatus" class="muted">Choose a folder.</small><button class="primary" type="button" id="pickerSelect" onclick="selectFolder()">Select folder</button></div></div></div>
<script>
const $=id=>document.getElementById(id);
let currentView='overview',pickerType='',folder='',parent='',folderName='',pickSeq=0;
const activePreviews=new Map();
function cancelPreview(id){
 const connection=activePreviews.get(id);
 if(connection){connection.close();activePreviews.delete(id)}
 const output=document.querySelector('[data-preview-result="'+id+'"]');
 if(output)output.textContent='Scan stopped. No files were changed.';
}
function renderPreviewResult(output,result){
  output.replaceChildren();
  const summary=document.createElement('div');summary.className='preview-summary';
  const totals=Object.entries(result.counts||{}).filter(([action,count])=>action!=='unchanged'&&count>0).sort((a,b)=>b[1]-a[1]);
  summary.textContent=totals.length?totals.map(([action,count])=>count+' '+action).join(' · '):'No changes detected';
  output.append(summary);
  const note=document.createElement('p');note.className='preview-notice';note.textContent='Preview only · No changes made · Generated '+new Date(result.generated).toLocaleString();output.append(note);
  const bootstrapInfo=document.createElement('p');bootstrapInfo.className='preview-notice';
  bootstrapInfo.textContent='Only actionable files are listed below. Unchanged files are excluded. For first-time matches, Syncria uses modification timestamps and file size without reading full file contents; ambiguity requires review.';
  output.append(bootstrapInfo);
  const filterBar=document.createElement('div');filterBar.className='preview-filter-bar';
  const filterTitle=document.createElement('label');filterTitle.textContent='Show files: ';
  const filterSelect=document.createElement('select');filterSelect.setAttribute('aria-label','Filter preview file actions');
  for(const [value,label] of [['all','All actions that matter'],['skip','Skipped and why'],['replace','Overwrites and why'],['copy','New copies'],['review','Needs a decision / blocked'],['delete','Deletes']]){
   const option=document.createElement('option');option.value=value;option.textContent=label;filterSelect.append(option);
  }
  filterTitle.append(filterSelect);filterBar.append(filterTitle);output.append(filterBar);
  const table=document.createElement('div');table.className='preview-table-wrap';
  const grid=document.createElement('table');grid.className='preview-table';
  const head=document.createElement('thead');
  const header=document.createElement('tr');
  for(const label of ['File or folder','Planned action','Direction / winner','Last modified (source / destination)','Why']){
   const th=document.createElement('th');th.textContent=label;header.append(th);
  }
  head.append(header);grid.append(head);
  const body=document.createElement('tbody');
  for(const row of result.rows||[]){
   const tr=document.createElement('tr');tr.dataset.action=row.action||'';
   const localTime=value=>value?new Date(value).toLocaleString():'Unknown';
   const modified=row.primary_modified||row.secondary_modified?
     'Primary: '+localTime(row.primary_modified)+' / Secondary: '+localTime(row.secondary_modified):'Not available';
   const displayAction=row.action;
   for(const value of [row.path,displayAction,(row.from&&row.to?row.from+' → '+row.to:row.from||''),modified,row.reason]){
    const td=document.createElement('td');td.textContent=value||'—';tr.append(td);
   }
   body.append(tr);
  }
  grid.append(body);table.append(grid);output.append(table);
  const setFilter=()=>{const choice=filterSelect.value;let shown=0;for(const tr of body.rows){const action=tr.dataset.action;const match=choice==='all'||action===choice||(choice==='review'&&(action==='blocked'||action==='review'));tr.hidden=!match;if(match)shown++}if(!shown){empty.textContent='No '+choice+' items were found.';empty.hidden=false}else{empty.hidden=true}};
  const empty=document.createElement('p');empty.className='muted';empty.hidden=true;output.append(empty);
  filterSelect.addEventListener('change',setFilter);setFilter();
  if(result.truncated){const p=document.createElement('p');p.className='preview-warning';p.textContent='The file list reached its 10,000 actionable-file limit. Counts include the complete scan.';output.append(p)}
  for(const warning of result.warnings||[]){const p=document.createElement('p');p.className='preview-notice';p.textContent=warning;output.append(p)}

}
function previewSync(id){
 const panel=document.querySelector('[data-preview="'+id+'"]');
 const output=document.querySelector('[data-preview-result="'+id+'"]');
 if(!panel||!output)return;
 cancelPreview(id);
 panel.classList.remove('hidden');
 output.replaceChildren();
 const status=document.createElement('div');status.className='preview-live-state';
 const heading=document.createElement('strong');heading.textContent='Scanning locations…';
 const description=document.createElement('div');description.textContent='Connecting to live scan. No files will be changed.';
 const counts=document.createElement('div');counts.className='preview-live-counts';
 const cancel=document.createElement('button');cancel.className='secondary';cancel.type='button';cancel.textContent='Stop scan';cancel.addEventListener('click',()=>cancelPreview(id));
 status.append(heading,description,counts,cancel);output.append(status);
 panel.scrollIntoView({behavior:'smooth',block:'nearest'});
 if(!('EventSource' in window)){
  description.textContent='Your browser does not support live scan streaming.';
  return;
 }
 const seen=new Map();
 const connection=new EventSource('/api/sync/preview/live?id='+encodeURIComponent(id));
 activePreviews.set(id,connection);
 connection.addEventListener('scan',event=>{
  if(activePreviews.get(id)!==connection)return;
  try{
   const scan=JSON.parse(event.data);
   if(scan.endpoint!=='Planning changes')seen.set(scan.endpoint,Number(scan.items||0));
   heading.textContent=scan.endpoint==='Planning changes'?'Comparing both locations…':'Scanning '+scan.endpoint;
   description.textContent=scan.endpoint==='Planning changes'?'Preparing the actionable file list…':'Reading file metadata (unchanged filenames are hidden)…';
   counts.textContent=[...seen].map(([name,count])=>name+': '+count.toLocaleString()+' items discovered').join(' · ')||'Starting scan…';
  }catch(_e){}
 });
 connection.addEventListener('complete',event=>{
  if(activePreviews.get(id)!==connection)return;
  connection.close();activePreviews.delete(id);
  try{renderPreviewResult(output,JSON.parse(event.data))}
  catch(error){output.textContent='Unable to show completed scan: '+error.message}
 });
 connection.addEventListener('failed',event=>{
  if(activePreviews.get(id)!==connection)return;
  connection.close();activePreviews.delete(id);
  try{output.textContent='Preview could not finish: '+(JSON.parse(event.data).error||'Unknown scan error')}
  catch(_e){output.textContent='Preview failed while reading scan results.'}
 });
 connection.onerror=()=>{
  if(activePreviews.get(id)!==connection)return;
  connection.close();activePreviews.delete(id);
  output.textContent='Live scan connection interrupted. No files were changed. Try again or check your reverse proxy streaming settings.';
 };
}
let recoveryDialogReturnFocus=null;
function openRecoveryDialog(id){
 const dialog=document.getElementById('recovery-'+id);
 if(!dialog||typeof dialog.showModal!=='function')return;
 // A native dialog must not remain beneath a hidden SPA view.
 if(dialog.parentElement!==document.body)document.body.append(dialog);
 recoveryDialogReturnFocus=document.activeElement;
 if(dialog.open)return;
 const input=document.getElementById('recovery-search-'+id);
 if(input)input.value='';
 refreshRecoveryExpirations();
 if(!dialog.querySelector('[data-recovery-path]'))return;
 dialog.showModal();
 if(input)input.focus();
}
function closeRecoveryDialog(id){
 const dialog=document.getElementById('recovery-'+id);
 if(dialog?.open)dialog.close();
}
function filterRecoveryFiles(id){
 const input=document.getElementById('recovery-search-'+id);
 const list=document.getElementById('recovery-items-'+id);
 if(!list)return;
 const term=(input?.value||'').trim().toLocaleLowerCase();
 const now=Date.now();
 let visible=0,total=0;
 for(const item of list.querySelectorAll('[data-recovery-path]')){
  const deadline=Number(item.dataset.recoveryExpiresMs);
  // Remove both the filename and its POST Restore form at expiration,
  // regardless of whether remote cleanup has succeeded yet.
  if(!Number.isFinite(deadline)||deadline<=now){
   item.remove();
   continue;
  }
  total++;
  const match=item.dataset.recoveryPath.toLocaleLowerCase().includes(term);
  item.hidden=!match;
  if(match)visible++;
 }
 const count=document.getElementById('recovery-visible-'+id);
 if(count)count.textContent=term?visible+' of '+total+' files shown':total+' files eligible for recovery';
 const summary=document.querySelector('[data-recovery-summary="'+CSS.escape(id)+'"]');
 if(summary){
  summary.hidden=total===0;
  const badge=summary.querySelector('.recovery-count');
  if(badge)badge.textContent=String(total);
 }
 const bar=document.getElementById('syncDetailRecoveryBar');
 if(bar&&selectedSyncID===id){
  bar.classList.toggle('hidden',total===0);
  const badge=document.getElementById('syncDetailRecoveryCount');
  if(badge)badge.textContent=String(total);
 }
 const empty=document.getElementById('recovery-empty-'+id);
 if(empty){
  empty.classList.toggle('hidden',visible!==0);
  empty.textContent=total===0?'The ten-minute recovery window has ended. No files remain eligible for restoration.':'No filenames match. Try another search.';
 }
}
let recoveryExpiryTimer=null;
function refreshRecoveryExpirations(){
 if(recoveryExpiryTimer!==null){clearTimeout(recoveryExpiryTimer);recoveryExpiryTimer=null}
 let nearest=Infinity;
 const now=Date.now();
 for(const dialog of document.querySelectorAll('dialog.recovery-dialog')){
  const id=dialog.id.slice('recovery-'.length);
  filterRecoveryFiles(id);
  for(const row of dialog.querySelectorAll('[data-recovery-expires-ms]')){
   const deadline=Number(row.dataset.recoveryExpiresMs);
   if(Number.isFinite(deadline)&&deadline>now)nearest=Math.min(nearest,deadline);
  }
 }
 if(Number.isFinite(nearest)){
  recoveryExpiryTimer=setTimeout(refreshRecoveryExpirations,Math.max(20,Math.min(nearest-Date.now()+20,60_000)));
 }
}
document.addEventListener('DOMContentLoaded',refreshRecoveryExpirations);
document.addEventListener('visibilitychange',()=>{if(!document.hidden)refreshRecoveryExpirations()});
window.addEventListener('focus',refreshRecoveryExpirations);
document.addEventListener('click',event=>{
 const modal=event.target;
 if(modal instanceof HTMLDialogElement&&modal.classList.contains('recovery-dialog'))modal.close();
});
document.addEventListener('close',event=>{
 if(!event.target?.classList?.contains('recovery-dialog'))return;
 const target=recoveryDialogReturnFocus;
 recoveryDialogReturnFocus=null;
 if(target?.isConnected&&typeof target.focus==='function')target.focus();
},true);
let selectedSyncID='',selectedSyncRunID='',runPageOffset=0,eventPageOffset=0;
let lastSyncDashboardJobs=[];
function makeHistoryNode(parent,tag,className,text){
 const n=document.createElement(tag);
 if(className)n.className=className;
 if(text!==undefined)n.textContent=text;
 parent.append(n);return n;
}
async function fetchSyncHistory(params){
 const url='/api/sync/history?'+new URLSearchParams(params);
 const result=await fetch(url,{credentials:'same-origin',cache:'no-store'});
 if(!result.ok)throw Error('History request failed (HTTP '+result.status+').');
 return await result.json();
}
function syncDurationText(run){
 if(!run.ended||!run.started)return 'In progress';
 const duration=Math.max(0,Date.parse(run.ended)-Date.parse(run.started));
 return (duration/1000).toFixed(1)+'s';
}
function formatHistoryDate(value){return value?new Date(value).toLocaleString():'Unknown time'}
function filterSyncLibrary(){
 const needle=($('syncLibrarySearch')?.value||'').trim().toLocaleLowerCase();
 const filter=$('syncLibraryFilter')?.value||'all';
 const cards=[...document.querySelectorAll('[data-sync-card]')];
 let shown=0,running=0,attention=0;
 for(const card of cards){
  const job=lastSyncDashboardJobs.find(j=>j.id===card.dataset.syncCard);
  const active=!!job?.running,problem=!!job?.error&&!active,enabled=job?.enabled??!!card.querySelector('.flow-state.is-on');
  if(active)running++;
  if(problem)attention++;
  const status=filter==='all'||filter==='running'&&active||filter==='attention'&&problem||
   filter==='active'&&enabled||filter==='paused'&&!enabled;
  const match=card.dataset.syncName.toLocaleLowerCase().includes(needle)&&status;
  card.hidden=!match;if(match)shown++;
 }
 const set=(id,value)=>{const el=$(id);if(el)el.textContent=String(value)};
 set('syncLibraryTotal',cards.length);set('syncLibraryRunning',running);set('syncLibraryAttention',attention);
 $('syncLibraryNoMatch')?.classList.toggle('hidden',shown!==0||cards.length===0);
}
document.addEventListener('click',e=>{
 for(const menu of document.querySelectorAll('details.sync-card-more[open]')){
  if(!menu.contains(e.target))menu.open=false;
 }
});
document.addEventListener('keydown',e=>{
 if(e.key==='Escape')document.querySelectorAll('details.sync-card-more[open]').forEach(menu=>menu.open=false);
});
let latestDetailedRun=null;
let mountedSyncPreview=null;
let selectedSyncDetailTab='overview';
function unmountSyncPreview(){
 if(!mountedSyncPreview)return;
 const prior=document.querySelector('[data-sync-card="'+CSS.escape(mountedSyncPreview)+'"]');
 const node=document.querySelector('[data-preview="'+CSS.escape(mountedSyncPreview)+'"]');
 if(prior&&node){
  const actions=prior.querySelector('.flow-card-actions');
  if(actions)prior.insertBefore(node,actions);
  node.classList.add('hidden');
 }
 mountedSyncPreview=null;
}
function mountSyncPreview(id){
 const host=$('syncDetailPreviewHost');if(!host)return;
 if(mountedSyncPreview===id&&host.querySelector('[data-preview="'+CSS.escape(id)+'"]'))return;
 unmountSyncPreview();
 host.replaceChildren();
 mountedSyncPreview=id;
 const preview=document.querySelector('[data-preview="'+CSS.escape(id)+'"]');
 if(preview){
  preview.classList.remove('hidden');
  const output=preview.querySelector('[data-preview-result]');
  if(output&&output.textContent==='Scanning both locations…')output.textContent='Run a read-only preview to see files Syncria would copy, replace, skip or delete.';
  host.append(preview);
 }else makeHistoryNode(host,'p','sync-detail-empty','Preview is unavailable for this connection.');
}
function showSyncDetailTab(name){
 const names=['overview','files','history','settings'];
 if(!names.includes(name))name='overview';
 selectedSyncDetailTab=name;
 if(currentView==='sync-detail'&&selectedSyncID)history.replaceState(null,'','#sync-detail/'+encodeURIComponent(selectedSyncID)+'/'+name);
 for(const tab of names){
  const key=tab[0].toUpperCase()+tab.slice(1);
  const b=$('syncTab'+key),panel=$('syncDetail'+key);
  if(b){b.setAttribute('aria-selected',String(tab===name));b.tabIndex=tab===name?0:-1}
  if(panel)panel.classList.toggle('is-active',tab===name);
 }
 if(name==='files'&&selectedSyncID)mountSyncPreview(selectedSyncID);
 // History is already loaded when the detail view opens.
}
function runDetailPreview(){
 if(!selectedSyncID)return;
 mountSyncPreview(selectedSyncID);previewSync(selectedSyncID);
}
function submitSyncAction(action){
 if(!selectedSyncID)return;
 if(action==='delete_job'&&!confirm('Remove this sync configuration? Files on your cloud and NAS will remain in place.'))return;
 if(action==='run_job'&&lastSyncDashboardJobs.find(j=>j.id===selectedSyncID)?.running)return;
 const form=document.createElement('form');form.method='post';form.action='/action';form.hidden=true;
 for(const [name,value] of Object.entries({action,id:selectedSyncID})){
  const input=document.createElement('input');input.type='hidden';input.name=name;input.value=value;form.append(input);
 }
 document.body.append(form);form.submit();
}
function renderLastDetailedRun(run){
 latestDetailedRun=run;
 const target=$('syncDetailLastRun');if(!target)return;
 target.className=run?'sync-detail-pair sync-detail-last-run':'sync-detail-empty';
 target.replaceChildren();
 if(!run){
  const empty=makeHistoryNode(target,'div','sync-detail-empty');
  makeHistoryNode(empty,'strong','', 'No detailed run history yet');
  makeHistoryNode(empty,'p','', 'New runs are recorded individually. Older release totals remain in Activity but cannot be reconstructed into a per-file journal.');
  return;
 }
 const top=makeHistoryNode(target,'div','sync-history-file-head');
 makeHistoryNode(top,'strong','',run.status==='success'?'Completed successfully':run.status==='failed'?'Finished with errors':run.status==='cancelled'?'Stopped':'In progress');
 makeHistoryNode(top,'span','',formatHistoryDate(run.started));
 makeHistoryNode(target,'div','sync-history-file-path',Number(run.event_count||0).toLocaleString()+' file operations · '+formatBytes(Number(run.upload_bytes||0)+Number(run.download_bytes||0))+' · '+syncDurationText(run));
 if(run.error)makeHistoryNode(target,'p','sync-detail-alert',run.error);
 const view=makeHistoryNode(target,'button','secondary','Inspect files in this run →');
 view.type='button';view.addEventListener('click',()=>{showSyncDetailTab('history');loadSyncRun(run.id)});
}
function renderSyncDetailLive(){
 if(!selectedSyncID)return;
 const job=lastSyncDashboardJobs.find(j=>j.id===selectedSyncID);
 const container=$('syncDetailLive');if(!container)return;
 if(!job){container.textContent='Connecting to live sync status…';return}
 const p=job.progress||{};
 const status=$('syncDetailStatus');
 const kind=job.running?'running':job.error?'attention':job.enabled?'healthy':'paused';
 status.dataset.kind=kind;
 status.textContent=job.running?'Syncing now':job.error?'Needs attention':job.enabled?'Automatic checks on':'Paused';
 const direction=job.provider==='nas_bridge'?(job.direction==='one_way'?'One-way · source NAS → destination NAS':'Two-way · NAS ↔ NAS'):job.direction==='one_way'?'One-way · cloud → NAS':'Two-way · cloud ↔ NAS';
 const deletion=job.propagate_deletions?'Deletions propagate to previously synchronized copies':'Deletions are not shared';
 const recovery=job.recovery_enabled?'10-minute deletion undo enabled':'Deletion undo disabled';
 const schedule=job.schedule_mode==='changes'?'On changes · five-minute checks':'Every configured interval';
 const recovered=document.querySelector('[data-recovery-summary="'+CSS.escape(selectedSyncID)+'"]');
 const recoveryBar=$('syncDetailRecoveryBar');
 if(recoveryBar)recoveryBar.classList.toggle('hidden',!recovered);
 const recoveries=$('syncDetailRecoveryCount');
 if(recoveries)recoveries.textContent=recovered?.querySelector('.recovery-count')?.textContent?.trim()||'';
 const watcher=job.schedule_mode==='changes'?(job.nas_watcher_verified?'NAS watcher verified; fast checks eligible':'NAS watcher unverified; a full metadata check may be required'):'Periodic scheduled checks';
 const recent=job.last_checked?'Last quick check '+job.last_checked:'No quick check recorded';
 const line=job.running?(p.phase==='scanning'?'Reading '+(p.current||'metadata')+' · '+Number(p.processed||0).toLocaleString()+' entries':p.phase==='transferring'?'Transferring '+(p.current||'file'):p.phase||'Working')+' · '+formatBytes(Number(p.upload_bytes||0)+Number(p.download_bytes||0))+' moved this run':recent+' · '+watcher;
 container.textContent=line;
 $('syncDetailSubtitle').textContent=direction+' · '+schedule;
 $('syncDetailDirection').textContent=direction;
 $('syncDetailDeletion').textContent=deletion;
 $('syncSettingDirection').textContent=direction;
 $('syncSettingSchedule').textContent=schedule;
 $('syncSettingDelete').textContent=deletion;
 $('syncSettingRecovery').textContent=recovery;
 $('syncDetailWatchNote').textContent=watcher+(job.last_run?' · Last full run '+job.last_run:'');
 $('syncDeleteWarning').textContent=job.propagate_deletions?
  'Deletion sharing is active. Removing a previously synchronized file from the source can also remove its counterpart, subject to recovery settings.':
  'Deletion sharing is off. Deleting a file at one location will not automatically delete its counterpart.';
 const set=(id,value)=>{const target=$(id);if(target)target.textContent=value};
 set('syncDetailSuccess',Number(job.successful_runs||0).toLocaleString());
 set('syncDetailMoved',(Number(job.total_uploads||0)+Number(job.total_downloads||0)).toLocaleString());
 set('syncDetailBytes',formatBytes(Number(job.total_upload_bytes||0)+Number(job.total_download_bytes||0)));
 const run=$('syncDetailRun'),stop=$('syncDetailStop'),pause=$('syncDetailPause');
 if(run)run.disabled=job.running;
 if(stop)stop.classList.toggle('hidden',!job.running);
 if(pause)pause.textContent=job.enabled?'Pause automatic checks':'Enable automatic checks';
}
function openSyncDetails(id,tab='overview'){
 if(!id)return;
 selectedSyncID=id;selectedSyncRunID='';runPageOffset=0;eventPageOffset=0;latestDetailedRun=null;
 const job=lastSyncDashboardJobs.find(j=>j.id===id);
 const card=document.querySelector('[data-sync-card="'+CSS.escape(id)+'"]');
 $('syncDetailTitle').textContent=job?.name||card?.dataset.syncName||'Sync details';
 $('syncRunSummary').textContent='Choose a run to inspect recorded files.';
 $('syncRunEvents').replaceChildren();$('syncRunList').textContent='Loading run history…';
 $('syncEventPager').replaceChildren();
 if($('syncFileSearch'))$('syncFileSearch').value='';
 if($('syncFileActionFilter'))$('syncFileActionFilter').value='';
 unmountSyncPreview();
 $('syncDetailPreviewHost').replaceChildren();
 renderLastDetailedRun(null);
 renderSyncDetailLive();navigate('sync-detail');
 showSyncDetailTab(tab);
 loadSyncRunList();
 if(!job)refreshDashboardStatus();
}
let runFileSearchTimer=null;
function debounceRunFileSearch(){
 if(runFileSearchTimer)clearTimeout(runFileSearchTimer);
 runFileSearchTimer=setTimeout(reloadRunFileFilters,300);
}
function reloadRunFileFilters(){
 if(selectedSyncRunID)loadSyncRun(selectedSyncRunID);
}
async function loadSyncRunList(more=false){
 if(!selectedSyncID)return;
 if(!more)runPageOffset=0;
 const jobID=selectedSyncID;
 const list=$('syncRunList'),pager=$('syncRunPager');
 if(!more){list.replaceChildren();makeHistoryNode(list,'p','muted','Loading run history…')}
 try{
  const data=await fetchSyncHistory({id:jobID,offset:runPageOffset});
  if(jobID!==selectedSyncID)return;
  if(!more)list.replaceChildren();
  if(!data.runs.length&&!more)makeHistoryNode(list,'p','muted','No detailed run records yet. Run history begins with this version; earlier runs only retained summary counts and up to 100 recent file events.');
  if(!more)renderLastDetailedRun(data.runs?.[0]||null);
  for(const run of data.runs||[]){
   const card=makeHistoryNode(list,'button','sync-history-run');
   card.type='button';card.dataset.runId=run.id;
   card.addEventListener('click',()=>loadSyncRun(run.id));
   makeHistoryNode(card,'strong','sync-history-run-date',formatHistoryDate(run.started));
   makeHistoryNode(card,'span','sync-history-run-status',run.status||'unknown');
   makeHistoryNode(card,'small','sync-history-run-stats',Number(run.uploads||0)+' → destination · '+Number(run.downloads||0)+' → source · '+formatBytes(Number(run.upload_bytes||0)+Number(run.download_bytes||0))+' · '+syncDurationText(run));
   makeHistoryNode(card,'small','sync-history-run-count',Number(run.event_count||0)+' file operations');
  }
  runPageOffset+=data.runs.length;
  pager.replaceChildren();
  if(data.more){
   const next=makeHistoryNode(pager,'button','secondary','Older runs');
   next.type='button';next.addEventListener('click',()=>loadSyncRunList(true));
  }
  if(!more&&data.runs.length)loadSyncRun(data.runs[0].id);
 }catch(error){list.textContent='Unable to load run history: '+error.message;}
}
async function loadSyncRun(runID,more=false){
 if(!selectedSyncID)return;
 if(!more){selectedSyncRunID=runID;eventPageOffset=0;}
 const id=selectedSyncID;
 if(!more){$('syncRunEvents').replaceChildren();$('syncRunSummary').textContent='Loading recorded file operations…'}
 try{
  const query=$('syncFileSearch')?.value?.trim()||'';
  const action=$('syncFileActionFilter')?.value||'';
  const data=await fetchSyncHistory({id,run:runID,offset:eventPageOffset,q:query,action});
  if(id!==selectedSyncID||runID!==selectedSyncRunID)return;
  for(const card of $('syncRunList').querySelectorAll('[data-run-id]'))card.classList.toggle('selected',card.dataset.runId===runID);
  const r=data.run,container=$('syncRunEvents'),pager=$('syncEventPager');
  if(!more)container.replaceChildren();
  $('syncRunSummary').textContent=formatHistoryDate(r.started)+' · '+r.status+' · '+(r.uploads||0)+' destination / '+(r.downloads||0)+' source copies · '+formatBytes(Number(r.upload_bytes||0)+Number(r.download_bytes||0))+' · '+syncDurationText(r)+(r.error?' · '+r.error:'');
  for(const event of data.events){
   const item=makeHistoryNode(container,'div','sync-history-file');
   const head=makeHistoryNode(item,'div','sync-history-file-head');
   makeHistoryNode(head,'strong','',event.action||'operation');
   makeHistoryNode(head,'span','',formatHistoryDate(event.at));
   makeHistoryNode(item,'div','sync-history-file-path',event.path||'');
   makeHistoryNode(item,'small','',event.from&&event.to?event.from+' → '+event.to:'');
   if(event.reason)makeHistoryNode(item,'small','sync-history-file-reason',event.reason);
  }
  eventPageOffset+=data.events.length;
  if(!eventPageOffset){
   const filtering=!!($('syncFileSearch')?.value||$('syncFileActionFilter')?.value);
   const empty=makeHistoryNode(container,'div','sync-detail-empty');
   makeHistoryNode(empty,'strong','',filtering?'No files match your filters':'No files needed changing');
   makeHistoryNode(empty,'p','',filtering?'Try a different filename or operation type. The entire run journal is searched.':'No file transfers were recorded for this run. Unchanged file paths are deliberately kept out of activity.');
  }
  pager.replaceChildren();
  if(data.more){
   const next=makeHistoryNode(pager,'button','secondary','Load more file operations');
   next.type='button';next.addEventListener('click',()=>loadSyncRun(runID,true));
  }
 }catch(error){$('syncRunSummary').textContent='Unable to load run details: '+error.message;}
}

let computerCache=[];
const computerSides={source:{id:'',root:'',path:'',parent:'',selected:null,seq:0},destination:{id:'',root:'',path:'',parent:'',selected:null,seq:0}};
async function computerJSON(url,options){
 const response=await fetch(url,{credentials:'same-origin',cache:'no-store',...options});
 if(!response.ok)throw Error((await response.text()).trim()||'HTTP '+response.status);
 return response.json();
}
async function pairComputer(){
 const button=$('computerPairButton'),message=$('computerPairMessage');
 button.disabled=true;message.textContent='Verifying encrypted connection…';
 try{
  await computerJSON('/api/computers/register',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({
   label:$('computerPairLabel').value,url:$('computerPairURL').value,key:$('computerPairKey').value,fingerprint:$('computerPairFingerprint').value
  })});
  $('computerPairForm').reset();message.textContent='Computer connected successfully.';
  await loadComputers();
 }catch(error){message.textContent=error.message}
 finally{button.disabled=false}
}
function computerOption(select,value,label){
 const option=document.createElement('option');option.value=value;option.textContent=label;select.append(option);
}
async function loadComputers(){
 let data;
 try{data=await computerJSON('/api/computers')}catch(error){$('computerDeviceList').textContent=error.message;return}
 computerCache=data.computers||[];
 const list=$('computerDeviceList');list.replaceChildren();
 if(!computerCache.length){const p=document.createElement('p');p.className='muted';p.textContent='No computers connected yet. Add one using the form.';list.append(p)}
 for(const c of computerCache){
  const row=document.createElement('div');row.className='computer-device';
  const desc=document.createElement('div'),name=document.createElement('strong'),meta=document.createElement('small'),status=document.createElement('small');
  name.textContent=c.label;meta.textContent=c.url+' · '+(c.roots||[]).length+' shared folders';
  status.textContent=c.online?'● Connected':'● Offline'+(c.error?' — '+c.error:'');
  status.className=c.online?'computer-online':'computer-offline';
  desc.append(name,meta,status);row.append(desc);
  const remove=document.createElement('button');remove.type='button';remove.className='secondary';remove.textContent='Remove';remove.addEventListener('click',()=>removeComputer(c));row.append(remove);list.append(row);
 }
 for(const side of ['source','destination']){
  const sel=$('computer'+(side==='source'?'Source':'Destination')+'ID'),old=sel.value;
  sel.replaceChildren();computerOption(sel,'','Choose computer');
  for(const c of computerCache)computerOption(sel,c.id,c.label+(c.online?'':' (offline)'));
  sel.value=computerCache.some(c=>c.id===old)?old:'';
  if(sel.value!==computerSides[side].id)computerSideChanged(side);
  else if(sel.value)computerUpdateRoots(side);
 }
}
async function removeComputer(c){
 if(!confirm('Remove '+c.label+' from Syncria? No files on the computer will be deleted.'))return;
 try{await computerJSON('/api/computers/remove',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id:c.id})});await loadComputers()}
 catch(error){alert(error.message)}
}
function computerSidePrefix(side){return side==='source'?'computerSource':'computerDestination'}
function computerUpdateRoots(side){
 const s=computerSides[side],prefix=computerSidePrefix(side);
 const dev=computerCache.find(c=>c.id===s.id),list=$(''+prefix+'Root'),old=list.value;
 list.replaceChildren();computerOption(list,'','Choose shared folder');
 for(const root of dev?.roots||[])computerOption(list,root.name,root.name);
 list.value=(dev?.roots||[]).some(x=>x.name===old)?old:'';
 if(s.root!==list.value)computerRootChanged(side);
}
function computerSideChanged(side){
 const s=computerSides[side],prefix=computerSidePrefix(side);
 s.id=$(prefix+'ID').value;s.root='';s.path='';s.selected=null;s.seq++;
 computerUpdateRoots(side);
 $(prefix+'Files').replaceChildren();$(prefix+'Path').textContent=s.id?'Choose a shared folder':'Choose a computer';
 if(side==='source')$('computerSelectedFile').textContent='No file selected';
}
function computerRootChanged(side){
 const s=computerSides[side],prefix=computerSidePrefix(side);
 s.root=$(prefix+'Root').value;s.path='';s.selected=null;s.seq++;
 if(side==='source')$('computerSelectedFile').textContent='No file selected';
 if(s.root)browseComputer(side,'');
 else{$(prefix+'Files').replaceChildren();$(prefix+'Path').textContent='Choose a shared folder'}
}
function computerUp(side){const s=computerSides[side];if(s.root&&s.path)browseComputer(side,s.parent)}
async function browseComputer(side,relative){
 const s=computerSides[side],prefix=computerSidePrefix(side),seq=++s.seq;
 const list=$(prefix+'Files');list.textContent='Loading files…';
 try{
  const query=new URLSearchParams({id:s.id,root:s.root,path:relative});
  const data=await computerJSON('/api/computers/browse?'+query);
  if(seq!==s.seq)return;
  s.path=data.path||'';s.parent=data.parent||'';
  if(side==='source'){s.selected=null;$('computerSelectedFile').textContent='No file selected'}
  $(prefix+'Path').textContent=s.root+(s.path?' / '+s.path:' /');
  list.replaceChildren();
  if(!(data.entries||[]).length){const p=document.createElement('p');p.className='muted';p.textContent='This folder is empty.';list.append(p)}
  for(const file of data.entries||[]){
   const row=document.createElement('button');row.type='button';row.className='computer-file';
   const name=document.createElement('strong');name.textContent=(file.directory?'📁 ':'📄 ')+file.name;
   const size=document.createElement('small');size.textContent=file.directory?'Open folder':formatBytes(file.size);
   row.append(name,size);
   if(file.directory)row.addEventListener('click',()=>browseComputer(side,s.path?s.path+'/'+file.name:file.name));
   else if(side==='source')row.addEventListener('click',()=>{
    list.querySelectorAll('.computer-file').forEach(node=>node.classList.remove('is-selected'));
    row.classList.add('is-selected');s.selected=file;
    $('computerSelectedFile').textContent=file.name+' ('+formatBytes(file.size)+')';
   });
   else{row.disabled=true;row.title='Choose a destination folder, not a file'}
   list.append(row);
  }
 }catch(error){if(seq===s.seq){list.textContent=error.message;$(prefix+'Path').textContent='Unable to browse shared folder'}}
}
async function startComputerTransfer(move){
 const from=computerSides.source,to=computerSides.destination;
 const message=$('computerTransferMessage');
 if(!from.selected||!from.id||!from.root||!to.id||!to.root){message.textContent='Choose a source file and destination shared folder.';return}
 if(from.id===to.id){message.textContent='Select two different computers.';return}
 const file=from.selected.name;
 const sourcePath=from.path?from.path+'/'+file:file;
 const destPath=to.path?to.path+'/'+file:file;
 if(!confirm((move?'Move':'Copy')+' '+file+' from '+computerCache.find(c=>c.id===from.id)?.label+' to '+computerCache.find(c=>c.id===to.id)?.label+'? Existing destination files will not be overwritten.'))return;
 message.textContent='Submitting secure direct transfer…';
 try{
  await computerJSON('/api/computers/transfers',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({
   source_id:from.id,destination_id:to.id,source_root:from.root,destination_root:to.root,
   source_path:sourcePath,destination_path:destPath,move
  })});
  message.textContent='Transfer queued. Progress appears below.';
  await refreshComputerTransfers();
 }catch(error){message.textContent='Transfer could not start: '+error.message}
}
async function refreshComputerTransfers(){
 let data;try{data=await computerJSON('/api/computers/transfers')}catch(error){$('computerTransferList').textContent=error.message;return}
 const list=$('computerTransferList');list.replaceChildren();
 if(!data.transfers?.length){const p=document.createElement('p');p.className='muted';p.textContent='No direct transfers yet.';list.append(p);return}
 for(const t of data.transfers){
  const row=document.createElement('div');row.className='computer-transfer-record';
  const name=document.createElement('strong');name.textContent=(t.move?'Move':'Copy')+': '+t.file+' — '+t.source+' → '+t.destination;
  const info=document.createElement('small');info.textContent=t.status+(t.bytes?' · '+formatBytes(t.bytes):'')+(t.error?' · '+t.error:'');
  row.append(name,info);list.append(row);
 }
}
setInterval(()=>{if(document.getElementById('view-computers')?.classList.contains('active'))refreshComputerTransfers()},2500);

function navigate(view){
 if(!$('view-'+view))return;
 const feedback=$('actionFeedback');if(feedback&&currentView!==view)feedback.remove();
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
  computers:['My computers','DIRECT CONNECTIONS'],
  accounts:['Cloud connections','CLOUD ACCOUNTS'],
  activity:['Transfer activity','ACTIVITY'],
  'sync-detail':['Sync details','SYNC DETAILS'],
  updates:['App updates','UPDATES'],
  settings:['Settings','PREFERENCES']
 };
 const page=pages[view]||[view,view.toUpperCase()];
 $('pageTitle').textContent=page[0];
 const breadcrumb=$('pageBreadcrumb');if(breadcrumb)breadcrumb.textContent=page[1];
 history.replaceState(null,'',view==='sync-detail'&&selectedSyncID?'#sync-detail/'+encodeURIComponent(selectedSyncID)+'/'+selectedSyncDetailTab:'#'+view);
 if(view==='updates')checkUpdate();
 if(view==='computers'){loadComputers();refreshComputerTransfers();}
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
 $('editingSyncId').value='';
 $('editingSyncType').value='';$('editingSyncType').disabled=true;
 $('relationshipType').disabled=false;
 $('editMappingWarning').classList.add('hidden');$('editResetMapping').checked=false;
 $('wizardModeLabel').textContent='New sync';$('wizardHeading').textContent='Connect your files.';
 $('createSyncButton').textContent='Save my sync (keep it off)';
 $('cloudDirection').value='bidirectional';
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
 $('nasDirection').disabled=cloud;
 $('cloudDirectionFields').classList.toggle('hidden',!cloud);
 $('cloudDirection').disabled=!cloud;
 $('cloudFirstRunFields').classList.toggle('hidden',!cloud);
 $('crossNASField').classList.toggle('hidden',!cloud);
 $('syncAccount').required=cloud&&$('sourceProvider').value==='google_drive';
 $('syncDropboxAccount').required=cloud&&$('sourceProvider').value==='dropbox';
 $('nasSourceId').required=!cloud;
 nasDirectionChanged();cloudDirectionChanged();
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
function cloudDirectionChanged(){
 const cloud=$('relationshipType').value==='cloud';
 const oneWay=cloud&&$('cloudDirection').value==='one_way';
 const cross=$('crossNASField');
 const checkbox=cross.querySelector('input[name="cross_nas_propagation"]');
 if(oneWay){checkbox.checked=false}
 checkbox.disabled=!cloud||oneWay;
 cross.classList.toggle('hidden',!cloud||oneWay);
 syncReviewed=false;
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
  ($('cloudDirection').value==='one_way'?'One way: cloud → NAS (cloud wins; NAS changes never upload)':'Both ways: cloud ↔ NAS')+
  ($('cloudDirection').value!=='one_way'&&$('crossNASField').querySelector('input').checked?' · Cross-NAS propagation enabled':'');
 $('reviewDeletion').textContent=($('propagateDeletions').checked?'Delete matching copies too':'Keep other copies if I delete a file')+
  ($('recoveryEnabled').checked?' · 10-minute Undo on':' · No Undo window');
 showSyncStep(4);syncReviewed=true;
}
function submitSyncWizard(event){
 event.preventDefault();if(syncStep!==4||!syncReviewed)return false;
 syncReviewed=false;$('createSyncButton').disabled=true;$('syncWizard').submit();return false;
}
async function editSync(id){
 try{
  const response=await fetch('/api/sync/config?id='+encodeURIComponent(id),{credentials:'same-origin',cache:'no-store'});
  if(!response.ok)throw Error((await response.text()).trim()||'Cannot load sync settings');
  const job=await response.json();
  startWizard();
  $('editingSyncId').value=job.id;
  $('editingSyncType').value=job.relationship_type;$('editingSyncType').disabled=false;
  $('relationshipType').value=job.relationship_type;
  $('relationshipType').disabled=true;
  $('wizardModeLabel').textContent='Edit sync';
  $('wizardHeading').textContent='Update '+job.name;
  $('createSyncButton').textContent='Save changes (keep paused)';
  $('editMappingWarning').classList.remove('hidden');
  $('syncName').value=job.name;
  $('sourceProvider').value=job.cloud_provider||'google_drive';
  relationshipTypeChanged();sourceProviderChanged();
  $('syncAccount').value=job.account_id||'';
  $('syncDropboxAccount').value=job.dropbox_account_id||'';
  $('remoteValue').value=job.remote_id||'root';
  $('selectedDriveLabel').textContent=job.remote_id==='root'?'Entire My Drive':(job.remote_id||'Selected cloud folder');
  $('nasSourceId').value=job.nas_source_id||'';
  $('nasSourceFolderValue').value=job.nas_source_folder||'.';
  $('selectedNASSourceLabel').textContent=job.nas_source_folder&&job.nas_source_folder!=='.'?job.nas_source_folder:'Share root';
  const destinationList=Array.isArray(job.nas_destinations)?job.nas_destinations:[];
  const choices=[...document.querySelectorAll('#nasTargets input[name="nas_ids"]')];
  for(const box of choices){
   const found=destinationList.find(d=>d.share_id===box.value||d.ShareID===box.value);
   box.checked=!!found;
   const folder=found?(found.folder??found.Folder??'.'):'.';
   const input=$('nas-folder-'+box.value);
   if(input)input.value=folder||'.';
   const caption=$('nas-label-'+box.value);
   if(caption)caption.textContent=!folder||folder==='.'?'Share root':folder;
  }
  $('cloudDirection').value=job.direction||'bidirectional';
  $('nasDirection').value=job.direction||'bidirectional';
  $('initialConflictPolicy').value=job.initial_conflict_policy||'newer';
  $('conflictPolicy').value=job.conflict_policy||'manual';
  $('scheduleMode').value=job.schedule_mode||'interval';
  scheduleModeChanged();
  $('syncInterval').value=String(job.interval||5);
  $('propagateDeletions').checked=!!job.propagate_deletions;
  $('recoveryEnabled').checked=!!job.recovery_enabled;
  $('crossNASField').querySelector('input').checked=!!job.cross_nas_propagation;
  cloudDirectionChanged();nasDirectionChanged();nasDestChanged();
  setWizardStatus(job.enabled?'Pause this sync before saving edits; the current settings remain unchanged until then.':'Edit the settings, verify both locations, and save. Nothing will be copied.');
 }catch(error){alert('Unable to edit this sync: '+(error?.message||'Unknown error'))}
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
// Only navigate after the new process confirms the expected version.
function finishVerifiedUpdate(version){
 if(!version || sessionStorage.getItem('syncria-update-refreshing')===version)return;
 sessionStorage.setItem('syncria-update-refreshing',version);
 sessionStorage.removeItem('syncria-update-expected');
 window.location.replace('/?updated='+encodeURIComponent(version)+'#updates');
}
function renderUpdateProgress(p){
 if(!p||!p.phase)return;
 const phase=p.phase;
 const expectedVersion=sessionStorage.getItem('syncria-update-expected');
 if(expectedVersion&&p.current===expectedVersion&&phase!=='failed'&&!p.active){
  finishVerifiedUpdate(expectedVersion);
  return;
 }
 if(phase==='complete'&&p.current&&(p.operation==='rollback'||p.operation==='manual'||p.target===p.current)){
  finishVerifiedUpdate(p.current);
  return;
 }
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
  const expected=sessionStorage.getItem('syncria-update-expected');
  if(expected&&payload.current===expected&&payload.phase!=='failed'&&!payload.active){finishVerifiedUpdate(expected);return;}
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
 $('updateError').classList.add('hidden');
 $('installUpdateButton').disabled=true;$('rollbackUpdateButton').disabled=true;
 try{
  const response=await fetch('/api/update/progress',{method:'POST',credentials:'same-origin',cache:'no-store',body:new URLSearchParams({operation})});
  if(!response.ok)throw Error((await response.text()).trim()||('HTTP '+response.status));
  const started=await response.json();
  if(started.target)sessionStorage.setItem('syncria-update-expected',started.target);
  sessionStorage.removeItem('syncria-update-refreshing');
  renderUpdateProgress(started);
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
 if($('updateStatusCard'))$('updateStatusCard').dataset.updateState='checking';
 if($('updateStatusHeadline'))$('updateStatusHeadline').textContent='Checking for updates…';
 if($('updateStatusIcon'))$('updateStatusIcon').textContent='↻';
 $('updateMessage').textContent='';updateInstallAvailable=false;
 $('installUpdateButton').disabled=true;
 try{
  const res=await fetch('/api/update/status'+(force?'?refresh=1':''),{credentials:'same-origin',cache:'no-store'});
  if(!res.ok)throw Error('Release check HTTP '+res.status);
  const result=await res.json();
  if(result.error)throw Error(result.error);
  $('latestVersion').textContent=result.latest||'No published release';
  $('updateAvailableVersion').textContent=result.latest||'None';
  const notes=$('updateReleaseNotes');
  if(notes){
   const valid=/^v?[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$/.test(result.latest||'');
   notes.href=valid?'https://github.com/gigabytegrove/syncria/releases/tag/'+encodeURIComponent(result.latest):'https://github.com/gigabytegrove/syncria/releases';
   notes.textContent=valid?'See what changed in '+result.latest+' ↗':'View published releases ↗';
  }
  const card=$('updateStatusCard'),headline=$('updateStatusHeadline'),icon=$('updateStatusIcon'),button=$('installUpdateButton');
  const state=result.available?'available':result.pending?'pending':'current';
  if(card)card.dataset.updateState=state;
  if(headline)headline.textContent=state==='available'?'Update available':state==='pending'?'Update is being prepared':'Syncria is up to date';
  if(icon)icon.textContent=state==='available'?'↓':state==='pending'?'…':'✓';
  $('updateMessage').textContent=state==='available'?'Syncria '+(result.latest||'')+' is ready to install on this device.':state==='pending'?(result.message||'The release files are not ready yet.'):'You’re running the latest published release.';
  if(button)button.textContent=state==='available'?'Download and install '+(result.latest||'update'):'No update available';
  updateInstallAvailable=!!result.available;
  if($('updateNavIndicator'))$('updateNavIndicator').classList.toggle('hidden',!result.available);
  $('installUpdateButton').disabled=updateRunning||updateSubmitting||!updateInstallAvailable;
 }catch(error){
  $('latestVersion').textContent='Release check failed';
  $('updateAvailableVersion').textContent='Unavailable';
  $('updateMessage').textContent=error.message;
  if($('installUpdateButton'))$('installUpdateButton').textContent='Update check unavailable';
  if($('updateStatusCard'))$('updateStatusCard').dataset.updateState='error';
  if($('updateStatusHeadline'))$('updateStatusHeadline').textContent='Unable to check for updates';
  if($('updateStatusIcon'))$('updateStatusIcon').textContent='!';
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

async function dismissSyncFailure(id,button){
 if(button){button.disabled=true;button.textContent='Dismissing…'}
 try{
  const response=await fetch('/action',{method:'POST',
   headers:{'Content-Type':'application/x-www-form-urlencoded'},
   body:new URLSearchParams({action:'dismiss_job_error',id}),
   credentials:'same-origin',cache:'no-store',redirect:'follow'});
  if(!response.ok)throw Error('Unable to dismiss the failure (HTTP '+response.status+').');
  await refreshDashboardStatus();
 }catch(error){
  alert(error.message);
 }finally{
  if(button){button.disabled=false;button.textContent='Dismiss old failure'}
 }
}
async function requestStopSync(id,button){
 if(!confirm('Stop this sync run? Files already copied will stay in place. Automatic checks will remain enabled unless you pause them separately.'))return;
 if(button){button.disabled=true;button.textContent='Stopping…'}
 try{
  const response=await fetch('/action',{method:'POST',body:new URLSearchParams({action:'cancel_job',id}),credentials:'same-origin',cache:'no-store'});
  if(!response.ok)throw Error('Stop request was not accepted (HTTP '+response.status+').');
  await refreshDashboardStatus();
 }catch(error){
  alert(error.message);
  if(button){button.disabled=false;button.textContent='Stop current run'}
 }
}
function liveText(parent,tag,className,value){
 const el=document.createElement(tag);
 if(className)el.className=className;
 if(value!==undefined)el.textContent=value;
 parent.append(el);return el;
}
function renderActiveSyncs(jobs){
 const root=$('activeSyncList');if(!root)return;
 const running=jobs.filter(job=>job.running);
 if(!running.length){
  if(root.dataset.running!=='none'){root.dataset.running='none';root.replaceChildren();liveText(root,'p','muted','No sync is running right now.')}
  return;
 }
 root.dataset.running='active';
 for(const child of [...root.children])if(!running.some(j=>j.id===child.dataset.runId))child.remove();
 for(const job of running){
  const p=job.progress||{};
  let card=[...root.children].find(el=>el.dataset.runId===job.id);
  if(!card){
   card=document.createElement('article');card.className='active-sync-card';card.dataset.runId=job.id;
   const heading=liveText(card,'div','active-sync-heading');
   const names=liveText(heading,'div','active-sync-title');
   liveText(names,'strong','active-sync-name',job.name||'Sync');
   liveText(names,'small','active-sync-stage','Preparing…');
   const stop=liveText(heading,'button','secondary active-sync-stop','Stop current run');stop.type='button';
   stop.addEventListener('click',()=>requestStopSync(job.id,stop));
   liveText(card,'p','active-sync-current','');
   const stats=liveText(card,'div','active-sync-stats');
   liveText(stats,'span','active-sync-scanned','');
   liveText(stats,'span','active-sync-count','');
   liveText(stats,'span','active-sync-bytes','');
   liveText(card,'p','active-sync-plan-label','Files waiting to sync');
   liveText(card,'div','active-sync-file-list');
   root.append(card);
  }
  const set=(q,value)=>{const el=card.querySelector(q);if(el&&el.textContent!==value)el.textContent=value};
  let phase=p.phase||'working';
  set('.active-sync-stage',phase==='scanning'?(p.initial_baseline?'Building initial file index':'Checking for changes'):phase==='stopping'?'Stopping safely…':phase==='transferring'?'Transferring files':phase==='processing'?'Checking changes':phase);
  const scanned=Number(p.processed||0),total=Number(p.total||0);
  set('.active-sync-scanned',phase==='scanning'?(p.current||'Location')+': '+scanned.toLocaleString()+' metadata entries':total>0?scanned.toLocaleString()+' / '+total.toLocaleString()+' entries checked':'Preparing file list');
  set('.active-sync-current',phase==='scanning'?(p.initial_baseline?'Building initial file index: ':'Checking ') +(p.current||'remote metadata')+' · '+Object.entries(p.scan_counts||{}).map(([k,v])=>k+': '+Number(v).toLocaleString()).join(' · '):p.current?'Current transfer: '+p.current:'Checking cloud/NAS changes against the local file index…');
  set('.active-sync-bytes',formatBytes(Number(p.upload_bytes||0)+Number(p.download_bytes||0))+' transferred this run');
  const recent=(p.events||[]).slice().reverse();
  const isTransferring=phase==='transferring'&&!!p.current;
  set('.active-sync-count',recent.length.toLocaleString()+' recent file operations');
  set('.active-sync-plan-label','Actual file operations (not a speculative transfer plan)');
  const signature=JSON.stringify({recent,current:p.current,transferring:isTransferring});
  if(card.dataset.planSignature!==signature){
   card.dataset.planSignature=signature;
   const list=card.querySelector('.active-sync-file-list');list.replaceChildren();
   if(isTransferring){
    const row=liveText(list,'div','active-sync-file is-active');
    const main=liveText(row,'div','active-sync-file-main');
    liveText(main,'strong','active-sync-file-action','Transferring');
    liveText(main,'span','active-sync-file-path',p.current);
    liveText(row,'small','active-sync-file-state','In progress');
   }
   for(const event of recent){
    if(!event.path)continue;
    const row=liveText(list,'div','active-sync-file is-done');
    const main=liveText(row,'div','active-sync-file-main');
    liveText(main,'strong','active-sync-file-action',event.action||'operation');
    const details=liveText(main,'div','active-sync-file-details');
    liveText(details,'span','active-sync-file-path',event.path);
    liveText(details,'small','active-sync-file-reason',(event.from||'')+(event.to?' → '+event.to:''));
    liveText(row,'small','active-sync-file-state','Recorded');
   }
   if(!recent.length&&!isTransferring)liveText(list,'p','muted','No file operations yet. Unchanged files are not listed. Use Test run / Preview for proposed changes.');
   if(recent.length){
    const all=liveText(list,'button','secondary','Open complete run history');
    all.type='button';all.addEventListener('click',()=>openSyncDetails(job.id));
   }
  }
 }
}

let dashboardPollBusy=false,dashboardSessionExpired=false;
function applyLiveStatus(data){
 if(!data||!Array.isArray(data.jobs))return;
 const jobs=data.jobs;
 lastSyncDashboardJobs=jobs;
 filterSyncLibrary();
 if(currentView==='sync-detail')renderSyncDetailLive();
 const active=jobs.filter(j=>j.running);
 renderActiveSyncs(jobs);
 const banner=$('liveSystemBar'),detail=$('liveSystemDetail'),dot=$('liveSystemDot');
 if(banner&&detail&&dot){
  banner.dataset.liveState=active.length?'running':'idle';
  dot.textContent=active.length?'●':'○';
  if(active.length){
   const current=active[0];
   const p=current.progress||{};
   detail.textContent=active.length+' sync'+(active.length===1?'':'s')+' running · '+(p.phase||'Working')+(p.current?' · '+p.current:'')+(Number(p.total||0)>0?' · '+Number(p.processed||0)+' / '+Number(p.total)+' checked':' · '+Number(p.processed||0)+' scanned');
  }else{
   const failures=jobs.filter(j=>j.error);
   detail.textContent=failures.length?failures.length+' sync'+(failures.length===1?' needs':'s need')+' attention':'All syncs idle · Live monitoring connected';
  }
 }
 const feed=$('liveActivityFeed');
 if(feed){
  const rows=[];
  for(const job of jobs){
   const events=(job.progress&&Array.isArray(job.progress.events))?job.progress.events:[];
   for(const event of events)rows.push({name:job.name||job.id,event});
  }
  rows.sort((x,y)=>Date.parse(y.event.at||0)-Date.parse(x.event.at||0));
  const newest=rows.slice(0,60);
  const signature=JSON.stringify(newest);
  if(feed.dataset.signature!==signature){
   feed.dataset.signature=signature;feed.replaceChildren();
   if(!newest.length){const p=document.createElement('p');p.className='muted';p.textContent='No file operations recorded yet. Changes will appear here while your syncs run.';feed.append(p);}
   for(const row of newest){
    const el=document.createElement('div');el.className='live-activity-item';
    const label=document.createElement('strong');label.textContent=row.event.action||'Changed';
    const filename=document.createElement('span');filename.textContent=row.event.path||'';
    const direction=document.createElement('small');direction.textContent=(row.event.from&&row.event.to?row.event.from+' → '+row.event.to:'')+' · '+row.name;
    el.append(label,filename,direction);feed.append(el);
   }
  }
 }
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
   const cancel=document.querySelector('[data-job-cancel="'+CSS.escape(job.id)+'"]');
   if(cancel)cancel.classList.toggle('hidden',!job.running);
   const run=document.querySelector('[data-job-run="'+CSS.escape(job.id)+'"]');
   if(run)run.textContent=job.last_run==='Never run'?'Never run':'Last run '+job.last_run;
   const summary=document.querySelector('[data-job-summary="'+CSS.escape(job.id)+'"]');
   if(summary)summary.textContent=job.summary||'';
   const error=document.querySelector('[data-job-error="'+CSS.escape(job.id)+'"]');
   if(error){error.hidden=!job.error||job.running;error.textContent=job.error?'Last sync failed ('+(job.last_run||'unknown date')+'): '+job.error:'';}
   const dismiss=document.querySelector('[data-job-dismiss="'+CSS.escape(job.id)+'"]');
   if(dismiss)dismiss.hidden=!job.error||job.running;
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
    const eventsBox=el('data-job-events'),eventsList=el('data-job-events-list');
    if(eventsBox&&eventsList){
      const events=Array.isArray(liveProgress.events)?liveProgress.events:[];
      eventsBox.classList.toggle('hidden',events.length===0);
      const signature=JSON.stringify(events);
      if(eventsList.dataset.signature!==signature){
        eventsList.dataset.signature=signature;eventsList.replaceChildren();
        for(const item of events.slice().reverse()){
          const row=document.createElement('div');row.className='sync-run-event';
          const action=document.createElement('strong');action.textContent=item.action||'Changed';
          const path=document.createElement('span');path.textContent=item.path||'';
          const direction=document.createElement('small');direction.textContent=item.from&&item.to?item.from+' → '+item.to:'';
          row.append(action,path,direction);eventsList.append(row);
        }
      }
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
}
let liveSource=null,liveFallback=null;
function connectLiveStatus(){
 if(!('EventSource' in window)){beginLiveFallback();return}
 try{
  liveSource=new EventSource('/api/live');
  liveSource.addEventListener('status',e=>{
   try{applyLiveStatus(JSON.parse(e.data));if(liveFallback){clearInterval(liveFallback);liveFallback=null}}
   catch(_e){}
  });
  liveSource.onerror=()=>{beginLiveFallback()};
 }catch(_e){beginLiveFallback()}
}
function beginLiveFallback(){
 if(liveFallback)return;
 refreshDashboardStatus();
 liveFallback=setInterval(refreshDashboardStatus,1500);
}
async function refreshDashboardStatus(){
 if(document.hidden||dashboardPollBusy||dashboardSessionExpired)return;
 dashboardPollBusy=true;
 try{
  const res=await fetch('/api/dashboard/status',{cache:'no-store',credentials:'same-origin'});
  if(res.status===401){dashboardSessionExpired=true;if(liveSource)liveSource.close();window.location.replace('/?msg=Session%20expired.%20Please%20sign%20in%20again.');return}
  if(res.ok)applyLiveStatus(await res.json());
 }catch(_e){}
 finally{dashboardPollBusy=false}
}
if(new URLSearchParams(location.search).has('updated')){
 sessionStorage.removeItem('syncria-update-refreshing');
 sessionStorage.removeItem('syncria-update-expected');
}
connectLiveStatus();
// Check quietly on startup so the nav can signal a verified available update.
checkUpdate();
setInterval(()=>{if(!document.hidden&&!updateRunning&&!updateSubmitting)checkUpdate()},30*60*1000);
// Live streaming is primary; polling starts only when unavailable.
document.addEventListener('visibilitychange',()=>{if(!document.hidden)refreshDashboardStatus()});
window.addEventListener('pageshow',event=>{if(event.persisted)refreshDashboardStatus();});
document.addEventListener('keydown',e=>{if(e.key==='Escape')closePicker()});
if(location.hash){
 const parts=location.hash.slice(1).split('/');
 if(parts[0]==='sync-detail'&&parts[1]){
  const id=decodeURIComponent(parts[1]);
  if(document.querySelector('[data-sync-card="'+CSS.escape(id)+'"]'))openSyncDetails(id,parts[2]||'overview');
  else navigate('sync');
 }else if(parts[0]==='sync'&&parts[1]==='recover'&&parts[2]){
  const id=decodeURIComponent(parts[2]);
  navigate('sync');
  openRecoveryDialog(id);
 }else if(parts[0]==='accounts'&&parts[1]==='connect'){
  startAccountWizard(parts[2]==='dropbox'?'dropbox':'google_drive');
 }else if($('view-'+parts[0])&&parts[0]!=='sync-detail')navigate(parts[0]);
}
</script>{{end}}</body></html>`
