<script>
  import { onMount } from 'svelte';
  
  onMount(() => {
    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn("backend-log", (msg) => {
        addLog(msg.trim());
      });
    }
  });

  // State management
  let currentPage = $state('scan');
  let sidebarCollapsed = $state(false);
  
  // Global config (shared across pages)
  let projectName = $state('default');
  let target = $state('');
  let depth = $state(1);
  let fastMode = $state(false);
  let recordVideo = $state(false);
  let mobileEmulation = $state('');
  let authJson = $state('');
  
  // Compare-specific
  let baselineDate = $state('');
  
  // Load-test-specific
  let users = $state(50);
  let duration = $state('10s');
  let method = $state('GET');
  let bodyJson = $state('');
  
  // Fuzz-API-specific
  let swaggerUrl = $state('');
  let fuzzConcurrency = $state(10);
  
  // Job status
  let jobStatus = $state('idle'); // idle | running | success | error
  let jobMessage = $state('');
  let jobLogs = $state([]);
  
  // History Explorer State
  let projects = $state([]);
  let selectedProject = $state('');
  let history = $state([]);
  let selectedRun = $state(null);
  let gallery = $state({ images: [], reports: [] });
  let loadingHistory = $state(false);
  
  // API base URL — same origin when embedded in Go
  const API_BASE = '/api';
  
  // Navigation items
  const navItems = [
    { id: 'scan', label: 'Scan', icon: '🔍', desc: 'Security Fuzzing' },
    { id: 'baseline', label: 'Baseline', icon: '📸', desc: 'Visual Snapshot' },
    { id: 'compare', label: 'Compare', icon: '🔬', desc: 'Visual Regression' },
    { id: 'a11y', label: 'A11y', icon: '♿', desc: 'Accessibility Audit' },
    { id: 'load', label: 'Load Test', icon: '🚀', desc: 'Stress Testing' },
    { id: 'fuzz-api', label: 'API Fuzzer', icon: '🔒', desc: 'Swagger Fuzzing' },
    { id: 'history', label: 'History', icon: '📂', desc: 'Results Explorer' },
  ];
  
  const mobileDevices = [
    '', 'iPhone 13', 'iPhone 14', 'iPhone 15', 'Pixel 5', 'Pixel 7',
    'Samsung Galaxy S21', 'iPad Mini', 'iPad Pro 11',
  ];
  
  function addLog(msg) {
    const ts = new Date().toLocaleTimeString();
    jobLogs = [...jobLogs, `[${ts}] ${msg}`];
  }
  
  async function executeJob(command) {
    if (!target) {
      jobStatus = 'error';
      jobMessage = 'Target URL is required!';
      return;
    }
    
    jobStatus = 'running';
    jobMessage = `Running ${command.toUpperCase()}...`;
    addLog(`▶ Starting ${command} against ${target}`);
    
    const payload = {
      project_name: projectName,
      command,
      target,
      depth,
      fast_mode: fastMode,
      record_video: recordVideo,
      mobile_emulation: mobileEmulation,
      auth_json: authJson,
    };
    
    // Add command-specific fields
    if (command === 'compare') {
      payload.baseline_date = baselineDate;
    }
    if (command === 'load') {
      payload.users = users;
      payload.duration = duration;
      payload.method = method;
      payload.body_json = bodyJson;
    }
    if (command === 'fuzz-api') {
      payload.swagger_url = swaggerUrl;
      payload.concurrency = fuzzConcurrency;
    }
    
    try {
      const res = await fetch(`${API_BASE}/run`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      
      const data = await res.json();
      
      if (res.ok) {
        jobStatus = 'success';
        jobMessage = data.message || `${command} completed!`;
        addLog(`✅ ${command} completed successfully.`);
      } else {
        jobStatus = 'error';
        jobMessage = data.error || 'An error occurred.';
        addLog(`❌ Error: ${data.error}`);
      }
    } catch (err) {
      jobStatus = 'error';
      jobMessage = `Connection failed: ${err.message}`;
      addLog(`❌ Connection to server failed: ${err.message}`);
    }
  }
  
  function resetStatus() {
    jobStatus = 'idle';
    jobMessage = '';
  }

  // History Functions
  async function loadProjects() {
    try {
      const res = await fetch(`${API_BASE}/projects`);
      if (res.ok) projects = await res.json();
    } catch (err) { console.error(err); }
  }

  async function loadHistory(proj) {
    selectedProject = proj;
    loadingHistory = true;
    selectedRun = null;
    gallery = { images: [], reports: [] };
    try {
      const res = await fetch(`${API_BASE}/history?project=${encodeURIComponent(proj)}`);
      if (res.ok) history = await res.json();
    } catch (err) { console.error(err); }
    loadingHistory = false;
  }

  async function viewRun(run) {
    selectedRun = run;
    try {
      const res = await fetch(`${API_BASE}/gallery?dir=${encodeURIComponent(run.report_dir)}`);
      if (res.ok) gallery = await res.json();
    } catch (err) { console.error(err); }
  }

  $effect(() => {
    if (currentPage === 'history') {
      loadProjects();
    }
  });
</script>

<!-- LAYOUT -->
<div class="flex min-h-screen bg-base-300">
  <!-- SIDEBAR -->
  <aside class="flex flex-col bg-base-200 border-r border-base-content/10 {sidebarCollapsed ? 'w-20' : 'w-72'} transition-all duration-300 ease-in-out">
    <!-- Logo -->
    <div class="flex items-center gap-3 px-5 py-5 border-b border-base-content/10">
      <div class="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-xl font-bold text-primary-content shadow-lg glow-primary">
        🕵️
      </div>
      {#if !sidebarCollapsed}
        <div class="animate-fade-in">
          <h1 class="text-sm font-bold text-base-content leading-tight">Web QA</h1>
          <p class="text-xs text-base-content/50">Automation Suite</p>
        </div>
      {/if}
    </div>
    
    <!-- Navigation -->
    <nav class="flex-1 py-4 px-3 space-y-1">
      {#each navItems as item}
        <button
          id="nav-{item.id}"
          class="w-full flex items-center gap-3 px-3 py-3 rounded-xl text-left transition-all duration-200
            {currentPage === item.id 
              ? 'bg-primary/15 text-primary border border-primary/20 shadow-sm glow-primary' 
              : 'text-base-content/70 hover:bg-base-content/5 hover:text-base-content border border-transparent'}"
          onclick={() => { currentPage = item.id; resetStatus(); }}
        >
          <span class="text-xl w-8 text-center flex-shrink-0">{item.icon}</span>
          {#if !sidebarCollapsed}
            <div class="animate-fade-in">
              <div class="text-sm font-semibold">{item.label}</div>
              <div class="text-xs opacity-50">{item.desc}</div>
            </div>
          {/if}
        </button>
      {/each}
    </nav>
    
    <!-- Collapse toggle -->
    <div class="p-3 border-t border-base-content/10">
      <button
        id="sidebar-toggle"
        class="w-full btn btn-ghost btn-sm text-base-content/50"
        onclick={() => sidebarCollapsed = !sidebarCollapsed}
      >
        {sidebarCollapsed ? '→' : '← Collapse'}
      </button>
    </div>
  </aside>

  <!-- MAIN CONTENT -->
  <main class="flex-1 flex flex-col overflow-hidden">
    
    <!-- HEADER -->
    <header class="flex items-center justify-between px-8 py-4 bg-base-200/50 backdrop-blur border-b border-base-content/10">
      <div>
        <h2 class="text-xl font-bold text-base-content">
          {navItems.find(n => n.id === currentPage)?.icon}
          {navItems.find(n => n.id === currentPage)?.label}
        </h2>
        <p class="text-sm text-base-content/50">{navItems.find(n => n.id === currentPage)?.desc}</p>
      </div>
      
      <!-- Status badge -->
      {#if jobStatus === 'running'}
        <div class="badge badge-warning gap-2 animate-pulse-glow py-3 px-4">
          <span class="loading loading-spinner loading-xs"></span>
          Running...
        </div>
      {:else if jobStatus === 'success'}
        <div class="badge badge-success gap-2 py-3 px-4">✅ Success</div>
      {:else if jobStatus === 'error'}
        <div class="badge badge-error gap-2 py-3 px-4">❌ Failed</div>
      {:else}
        <div class="badge badge-ghost gap-2 py-3 px-4">⏸️ Ready</div>
      {/if}
    </header>
    
    <!-- PAGE CONTENT -->
    <div class="flex-1 overflow-y-auto p-8">
      
      {#if currentPage !== 'history' && currentPage !== 'fuzz-api'}
      <!-- SHARED CONFIG CARD (Global) -->
      <div class="card bg-base-200 shadow-xl border border-base-content/5 mb-6 animate-fade-in">
        <div class="card-body">
          <h3 class="card-title text-base-content text-lg mb-4">
            ⚙️ General Configuration
          </h3>
          
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <!-- Project Name -->
            <div class="form-control col-span-1 md:col-span-1">
              <label class="label" for="input-project">
                <span class="label-text font-semibold">📁 Project Name</span>
              </label>
              <input
                id="input-project"
                type="text"
                placeholder="default"
                class="input input-bordered input-primary w-full"
                bind:value={projectName}
              />
            </div>

            <!-- Target URL -->
            <div class="form-control col-span-1 md:col-span-1 lg:col-span-2">
              <label class="label" for="input-target">
                <span class="label-text font-semibold">🎯 Target URL <span class="text-error">*</span></span>
              </label>
              <input
                id="input-target"
                type="url"
                placeholder="https://example.com"
                class="input input-bordered input-primary w-full font-mono text-sm focus:glow-primary"
                bind:value={target}
              />
            </div>
            
            <!-- Depth -->
            <div class="form-control">
              <label class="label" for="input-depth">
                <span class="label-text font-semibold">📏 Crawl Depth</span>
                <span class="label-text-alt badge badge-primary badge-sm">{depth}</span>
              </label>
              <input
                id="input-depth"
                type="range"
                min="1" max="5"
                class="range range-primary range-sm"
                bind:value={depth}
              />
              <div class="flex justify-between px-1 mt-1">
                {#each [1,2,3,4,5] as v}
                  <span class="text-xs text-base-content/40">{v}</span>
                {/each}
              </div>
            </div>
            
            <!-- Mobile Emulation -->
            <div class="form-control">
              <label class="label" for="input-mobile">
                <span class="label-text font-semibold">📱 Mobile Emulation</span>
              </label>
              <select id="input-mobile" class="select select-bordered select-sm w-full" bind:value={mobileEmulation}>
                <option value="">Desktop (Default)</option>
                {#each mobileDevices.filter(d => d) as device}
                  <option value={device}>{device}</option>
                {/each}
              </select>
            </div>
            
            <!-- Auth JSON -->
            <div class="form-control">
              <label class="label" for="input-auth">
                <span class="label-text font-semibold">🔐 Auth JSON Path</span>
              </label>
              <input
                id="input-auth"
                type="text"
                placeholder="./session.json"
                class="input input-bordered input-sm w-full font-mono"
                bind:value={authJson}
              />
            </div>
          </div>
          
          <!-- Toggles -->
          <div class="flex flex-wrap gap-6 mt-4 pt-4 border-t border-base-content/5">
            <label class="label cursor-pointer gap-3" for="toggle-fast">
              <span class="label-text text-sm">⚡ Fast Mode</span>
              <input id="toggle-fast" type="checkbox" class="toggle toggle-primary toggle-sm" bind:checked={fastMode} />
            </label>
            <label class="label cursor-pointer gap-3" for="toggle-video">
              <span class="label-text text-sm">🎥 Record Video</span>
              <input id="toggle-video" type="checkbox" class="toggle toggle-secondary toggle-sm" bind:checked={recordVideo} />
            </label>
          </div>
        </div>
      </div>
      {/if}
      
      <!-- COMMAND-SPECIFIC CARDS -->
      
      <!-- SCAN PAGE -->
      {#if currentPage === 'scan'}
        <div class="card bg-base-200 shadow-xl border border-base-content/5 animate-fade-in">
          <div class="card-body">
            <h3 class="card-title text-base-content">
              🔍 Security Scan & Fuzzing
            </h3>
            <p class="text-sm text-base-content/60 mb-4">
              The bot will crawl the target, intercept network traffic, perform Form Injection, Rage Clicks, and JWT Token Tampering. 
              All network anomalies will be automatically logged.
            </p>
            
            <div class="flex flex-wrap gap-3">
              <div class="stats stats-horizontal bg-base-300 shadow border border-base-content/5">
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Mode</div>
                  <div class="stat-value text-sm text-primary">Full Audit</div>
                </div>
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Fuzzing</div>
                  <div class="stat-value text-sm text-warning">Active</div>
                </div>
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Interception</div>
                  <div class="stat-value text-sm text-success">Active</div>
                </div>
              </div>
            </div>

            <div class="card-actions justify-end mt-6">
              <button
                id="btn-scan"
                class="btn btn-primary btn-wide gap-2 shadow-lg hover:glow-primary"
                onclick={() => executeJob('scan')}
                disabled={jobStatus === 'running'}
              >
                {#if jobStatus === 'running' && currentPage === 'scan'}
                  <span class="loading loading-spinner loading-sm"></span>
                {/if}
                🚀 Start Scan
              </button>
            </div>
          </div>
        </div>
        
      <!-- BASELINE PAGE -->
      {:else if currentPage === 'baseline'}
        <div class="card bg-base-200 shadow-xl border border-base-content/5 animate-fade-in">
          <div class="card-body">
            <h3 class="card-title text-base-content">
              📸 Take Baseline Screenshot
            </h3>
            <p class="text-sm text-base-content/60 mb-4">
              The bot will "walk politely" (without fuzzing) to take clean screenshots of each page. 
              Results are saved in <code class="text-primary text-xs">proofs/&lt;project_name&gt;/&lt;timestamp&gt;_baseline/</code>
            </p>
            
            <div class="alert alert-info shadow-lg">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
              <span class="text-sm">Read-Only Mode: Safe to run even on Production servers.</span>
            </div>

            <div class="card-actions justify-end mt-6">
              <button
                id="btn-baseline"
                class="btn btn-accent btn-wide gap-2 shadow-lg"
                onclick={() => executeJob('baseline')}
                disabled={jobStatus === 'running'}
              >
                {#if jobStatus === 'running' && currentPage === 'baseline'}
                  <span class="loading loading-spinner loading-sm"></span>
                {/if}
                📸 Take Snapshot
              </button>
            </div>
          </div>
        </div>
        
      <!-- COMPARE PAGE -->
      {:else if currentPage === 'compare'}
        <div class="card bg-base-200 shadow-xl border border-base-content/5 animate-fade-in">
          <div class="card-body">
            <h3 class="card-title text-base-content">
              🔬 Visual Regression Compare
            </h3>
            <p class="text-sm text-base-content/60 mb-4">
              Compare the current website appearance with a previous baseline. 
              Changed areas will be highlighted in bright red in an "X-ray" image.
            </p>
            
            <div class="form-control max-w-sm">
              <label class="label" for="input-baseline-date">
                <span class="label-text font-semibold">📅 Comparison Baseline Folder Path</span>
              </label>
              <input
                id="input-baseline-date"
                type="text"
                placeholder="proofs/my-project/2026-01-01_12-00-00_baseline"
                class="input input-bordered input-sm w-full font-mono"
                bind:value={baselineDate}
              />
              <label class="label">
                <span class="label-text-alt text-base-content/40 text-xs">Use "History Explorer" to find your baseline path.</span>
              </label>
            </div>

            <div class="card-actions justify-end mt-6">
              <button
                id="btn-compare"
                class="btn btn-secondary btn-wide gap-2 shadow-lg"
                onclick={() => executeJob('compare')}
                disabled={jobStatus === 'running'}
              >
                {#if jobStatus === 'running' && currentPage === 'compare'}
                  <span class="loading loading-spinner loading-sm"></span>
                {/if}
                🔬 Start Comparison
              </button>
            </div>
          </div>
        </div>
        
      <!-- A11Y PAGE -->
      {:else if currentPage === 'a11y'}
        <div class="card bg-base-200 shadow-xl border border-base-content/5 animate-fade-in">
          <div class="card-body">
            <h3 class="card-title text-base-content">
              ♿ Accessibility Audit (WCAG)
            </h3>
            <p class="text-sm text-base-content/60 mb-4">
              Inject axe-core into every page to detect accessibility violations: 
              color contrast, ARIA labels, keyboard navigation, and more.
            </p>
            
            <div class="flex flex-wrap gap-3">
              <div class="stats stats-horizontal bg-base-300 shadow border border-base-content/5">
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Engine</div>
                  <div class="stat-value text-sm text-info">axe-core</div>
                </div>
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Standard</div>
                  <div class="stat-value text-sm text-warning">WCAG 2.1</div>
                </div>
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Dynamic UI</div>
                  <div class="stat-value text-sm text-success">Yes</div>
                </div>
              </div>
            </div>

            <div class="card-actions justify-end mt-6">
              <button
                id="btn-a11y"
                class="btn btn-info btn-wide gap-2 shadow-lg"
                onclick={() => executeJob('a11y')}
                disabled={jobStatus === 'running'}
              >
                {#if jobStatus === 'running' && currentPage === 'a11y'}
                  <span class="loading loading-spinner loading-sm"></span>
                {/if}
                ♿ Start Audit
              </button>
            </div>
          </div>
        </div>
        
      <!-- LOAD TEST PAGE -->
      {:else if currentPage === 'load'}
        <div class="card bg-base-200 shadow-xl border border-base-content/5 animate-fade-in">
          <div class="card-body">
            <h3 class="card-title text-base-content">
              🚀 Load & Stress Testing
            </h3>
            <p class="text-sm text-base-content/60 mb-4">
              Shoot thousands of simultaneous HTTP requests at the target using Goroutines. 
              Supports POST with dynamic payload <code class="text-primary text-xs">{'{{RANDOM}}'}</code>.
            </p>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Users -->
              <div class="form-control">
                <label class="label" for="input-users">
                  <span class="label-text font-semibold">👥 Virtual Users</span>
                  <span class="label-text-alt badge badge-primary badge-sm">{users}</span>
                </label>
                <div class="join w-full">
                  <input
                    id="input-users-range"
                    type="range"
                    min="10" max="5000" step="10"
                    class="range range-primary range-sm join-item"
                    bind:value={users}
                  />
                  <input
                    id="input-users"
                    type="number"
                    min="1"
                    class="input input-bordered input-sm w-24 font-mono join-item"
                    bind:value={users}
                  />
                </div>
                <div class="flex justify-between px-1 mt-1">
                  <span class="text-[10px] text-base-content/40">10</span>
                  <span class="text-[10px] text-base-content/40">2500</span>
                  <span class="text-[10px] text-base-content/40">5000+</span>
                </div>
                {#if users > 1000}
                  <span class="text-[10px] text-warning mt-1 italic">⚠️ High concurrency may be limited by your local OS (ulimit/sockets).</span>
                {/if}
              </div>
              
              <!-- Duration -->
              <div class="form-control">
                <label class="label" for="input-duration">
                  <span class="label-text font-semibold">⏱️ Duration</span>
                </label>
                <input
                  id="input-duration"
                  type="text"
                  placeholder="10s, 1m, 2m30s"
                  class="input input-bordered input-sm w-full font-mono"
                  bind:value={duration}
                />
              </div>
              
              <!-- Method -->
              <div class="form-control">
                <label class="label" for="input-method">
                  <span class="label-text font-semibold">📡 HTTP Method</span>
                </label>
                <select id="input-method" class="select select-bordered select-sm w-full" bind:value={method}>
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                </select>
              </div>
              
              <!-- Body JSON -->
              <div class="form-control">
                <label class="label" for="input-bodyjson">
                  <span class="label-text font-semibold">📦 Body JSON</span>
                </label>
                <input
                  id="input-bodyjson"
                  type="text"
                  placeholder={'{"email": "test_{{RANDOM}}@mail.com"}'}
                  class="input input-bordered input-sm w-full font-mono text-xs"
                  bind:value={bodyJson}
                />
              </div>
            </div>

            <div class="alert alert-warning shadow-lg mt-4">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
              <span class="text-sm">Warning: Do not run load tests on production servers without permission!</span>
            </div>

            <div class="card-actions justify-end mt-6">
              <button
                id="btn-load"
                class="btn btn-warning btn-wide gap-2 shadow-lg"
                onclick={() => executeJob('load')}
                disabled={jobStatus === 'running'}
              >
                {#if jobStatus === 'running' && currentPage === 'load'}
                  <span class="loading loading-spinner loading-sm"></span>
                {/if}
                ⚡ Start Load Test
              </button>
            </div>
          </div>
        </div>

      <!-- API FUZZER PAGE -->
      {:else if currentPage === 'fuzz-api'}
        <div class="card bg-base-200 shadow-xl border border-base-content/5 animate-fade-in">
          <div class="card-body">
            <h3 class="card-title text-base-content">
              🔒 Swagger/OpenAPI Smart Fuzzer
            </h3>
            <p class="text-sm text-base-content/60 mb-4">
              Provide a Swagger/OpenAPI spec URL or file path. The engine will automatically parse all endpoints,
              generate type-aware payloads (SQLi, XSS, boundary, type confusion), and fire them concurrently.
            </p>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Project Name -->
              <div class="form-control">
                <label class="label" for="input-fuzz-project">
                  <span class="label-text font-semibold">📁 Project Name</span>
                </label>
                <input
                  id="input-fuzz-project"
                  type="text"
                  placeholder="default"
                  class="input input-bordered input-primary w-full"
                  bind:value={projectName}
                />
              </div>

              <!-- Target API Base URL -->
              <div class="form-control">
                <label class="label" for="input-fuzz-target">
                  <span class="label-text font-semibold">🎯 API Base URL <span class="text-error">*</span></span>
                </label>
                <input
                  id="input-fuzz-target"
                  type="url"
                  placeholder="https://api.example.com"
                  class="input input-bordered input-primary w-full font-mono text-sm"
                  bind:value={target}
                />
              </div>
              
              <!-- Swagger URL -->
              <div class="form-control md:col-span-2">
                <label class="label" for="input-swagger-url">
                  <span class="label-text font-semibold">📄 Swagger/OpenAPI Spec URL <span class="text-error">*</span></span>
                </label>
                <input
                  id="input-swagger-url"
                  type="text"
                  placeholder="https://api.example.com/swagger.json or ./openapi.yaml"
                  class="input input-bordered input-accent w-full font-mono text-sm focus:glow-primary"
                  bind:value={swaggerUrl}
                />
                <label class="label">
                  <span class="label-text-alt text-base-content/40 text-xs">Supports Swagger 2.0, OpenAPI 3.x (JSON or YAML)</span>
                </label>
              </div>
              
              <!-- Concurrency -->
              <div class="form-control">
                <label class="label" for="input-fuzz-concurrency">
                  <span class="label-text font-semibold">⚡ Concurrency</span>
                  <span class="label-text-alt badge badge-primary badge-sm">{fuzzConcurrency} workers</span>
                </label>
                <input
                  id="input-fuzz-concurrency"
                  type="range"
                  min="1" max="50"
                  class="range range-primary range-sm"
                  bind:value={fuzzConcurrency}
                />
                <div class="flex justify-between px-1 mt-1">
                  <span class="text-[10px] text-base-content/40">1</span>
                  <span class="text-[10px] text-base-content/40">25</span>
                  <span class="text-[10px] text-base-content/40">50</span>
                </div>
              </div>
            </div>
            
            <div class="flex flex-wrap gap-3 mt-4">
              <div class="stats stats-horizontal bg-base-300 shadow border border-base-content/5">
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Engine</div>
                  <div class="stat-value text-sm text-primary">kin-openapi</div>
                </div>
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Payloads</div>
                  <div class="stat-value text-sm text-warning">Smart Gen</div>
                </div>
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Vectors</div>
                  <div class="stat-value text-sm text-error">SQLi/XSS/RCE</div>
                </div>
                <div class="stat px-4 py-3">
                  <div class="stat-title text-xs">Execution</div>
                  <div class="stat-value text-sm text-success">Concurrent</div>
                </div>
              </div>
            </div>

            <div class="alert alert-warning shadow-lg mt-4">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
              <span class="text-sm">This tool sends security payloads. Only use on APIs you have permission to test!</span>
            </div>

            <div class="card-actions justify-end mt-6">
              <button
                id="btn-fuzz-api"
                class="btn btn-error btn-wide gap-2 shadow-lg hover:glow-error"
                onclick={() => executeJob('fuzz-api')}
                disabled={jobStatus === 'running' || !swaggerUrl || !target}
              >
                {#if jobStatus === 'running' && currentPage === 'fuzz-api'}
                  <span class="loading loading-spinner loading-sm"></span>
                {/if}
                🔒 Start API Fuzzing
              </button>
            </div>
          </div>
        </div>

      <!-- HISTORY EXPLORER PAGE -->
      {:else if currentPage === 'history'}
        <div class="grid grid-cols-1 lg:grid-cols-4 gap-6 animate-fade-in">
          <!-- Project List -->
          <div class="lg:col-span-1 space-y-4">
            <div class="card bg-base-200 shadow border border-base-content/5">
              <div class="card-body p-4">
                <h3 class="font-bold text-sm mb-2 uppercase tracking-wider text-base-content/50">Projects</h3>
                <div class="space-y-1">
                  {#each projects as p}
                    <button 
                      class="btn btn-sm w-full justify-start {selectedProject === p ? 'btn-primary' : 'btn-ghost'}"
                      onclick={() => loadHistory(p)}
                    >
                      📁 {p}
                    </button>
                  {:else}
                    <div class="text-xs opacity-50 p-2 text-center">No projects found.</div>
                  {/each}
                </div>
              </div>
            </div>
          </div>

          <!-- History List & Details -->
          <div class="lg:col-span-3 space-y-6">
            {#if selectedProject}
              <div class="card bg-base-200 shadow border border-base-content/5">
                <div class="card-body p-4">
                  <h3 class="font-bold text-lg mb-4">📜 History: {selectedProject}</h3>
                  
                  {#if loadingHistory}
                    <div class="flex justify-center p-8"><span class="loading loading-spinner loading-lg"></span></div>
                  {:else}
                    <div class="overflow-x-auto">
                      <table class="table table-sm w-full">
                        <thead>
                          <tr>
                            <th>Type</th>
                            <th>Timestamp</th>
                            <th>URL</th>
                            <th>Status</th>
                            <th>Action</th>
                          </tr>
                        </thead>
                        <tbody>
                          {#each history as run}
                            <tr class={selectedRun?.id === run.id ? 'bg-base-300' : ''}>
                              <td><div class="badge badge-outline badge-xs">{run.test_type}</div></td>
                              <td class="text-xs font-mono">{run.timestamp}</td>
                              <td class="text-xs truncate max-w-[150px]">{run.target_url}</td>
                              <td>
                                <div class="badge {run.status === 'completed' ? 'badge-success' : 'badge-error'} badge-xs">
                                  {run.status}
                                </div>
                              </td>
                              <td>
                                <button class="btn btn-xs btn-ghost" onclick={() => viewRun(run)}>👁️ View</button>
                              </td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  {/if}
                </div>
              </div>

              <!-- Gallery View -->
              {#if selectedRun}
                <div class="card bg-base-200 shadow border border-base-content/5 animate-fade-in">
                  <div class="card-body p-6">
                    <div class="flex justify-between items-start mb-6">
                      <div>
                        <h3 class="font-bold text-xl">🖼️ Gallery & Results</h3>
                        <p class="text-xs opacity-50 font-mono mt-1">{selectedRun.report_dir}</p>
                      </div>
                      <div class="flex gap-2">
                        {#if selectedRun.test_type === 'baseline'}
                          <button 
                            class="btn btn-xs btn-outline btn-secondary"
                            onclick={() => { 
                              baselineDate = selectedRun.report_dir; 
                              currentPage = 'compare';
                            }}
                          >
                            Set as Baseline for Compare
                          </button>
                        {/if}
                      </div>
                    </div>

                    <!-- Images Grid -->
                    {#if gallery.images.length > 0}
                      <h4 class="font-bold text-sm mb-3">Screenshots</h4>
                      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                        {#each gallery.images as img}
                          <div class="group relative rounded-lg overflow-hidden border border-base-content/10 shadow-sm aspect-video bg-base-300">
                            <img 
                              src="/{selectedRun.report_dir}/{img}" 
                              alt={img}
                              class="w-full h-full object-cover hover:scale-105 transition-transform duration-300 cursor-pointer"
                              onclick={() => window.open(`/${selectedRun.report_dir}/${img}`, '_blank')}
                            />
                            <div class="absolute bottom-0 left-0 right-0 bg-black/60 p-1 text-[10px] text-white opacity-0 group-hover:opacity-100 transition-opacity truncate">
                              {img}
                            </div>
                          </div>
                        {/each}
                      </div>
                    {/if}

                    <!-- Reports List -->
                    {#if gallery.reports.length > 0}
                      <h4 class="font-bold text-sm mb-3 mt-6">Generated Reports</h4>
                      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {#each gallery.reports as rpt}
                          <div class="flex items-center gap-3 p-3 bg-base-300 rounded-lg border border-base-content/5">
                            <span class="text-xl">📄</span>
                            <div class="flex-1 overflow-hidden">
                              <div class="text-sm font-semibold truncate">{rpt}</div>
                              <div class="text-[10px] opacity-40 uppercase">{rpt.split('.').pop()}</div>
                            </div>
                            <a 
                              href="/{selectedRun.report_dir}/{rpt}" 
                              target="_blank" 
                              class="btn btn-xs btn-primary"
                            >
                              Open
                            </a>
                          </div>
                        {/each}
                      </div>
                    {/if}

                    {#if gallery.images.length === 0 && gallery.reports.length === 0}
                      <div class="alert alert-info">No visual proofs found in this directory.</div>
                    {/if}
                  </div>
                </div>
              {/if}
            {:else}
              <div class="card bg-base-200 shadow border border-base-content/5">
                <div class="card-body items-center justify-center p-20 text-center opacity-30">
                  <div class="text-6xl mb-4">📂</div>
                  <div class="font-bold text-xl">Select a project to view history</div>
                  <p class="text-sm">Each test run will be automatically recorded.</p>
                </div>
              </div>
            {/if}
          </div>
        </div>
      {/if}
      
      <!-- STATUS & LOGS CARD -->
      {#if jobMessage}
        <div class="card bg-base-200 shadow-xl border border-base-content/5 mt-6 animate-fade-in">
          <div class="card-body">
            <h3 class="card-title text-base-content text-lg">
              📋 Status & Log
            </h3>
            
            <!-- Status Message -->
            <div class="alert {jobStatus === 'success' ? 'alert-success' : jobStatus === 'error' ? 'alert-error' : 'alert-warning'} shadow-sm">
              <span>{jobMessage}</span>
            </div>
            
            <!-- Logs -->
            {#if jobLogs.length > 0}
              <div class="bg-base-300 rounded-xl p-4 mt-3 max-h-60 overflow-y-auto font-mono text-xs">
                {#each jobLogs as log}
                  <div class="py-0.5 text-base-content/70">{log}</div>
                {/each}
              </div>
            {/if}
            
            <!-- Clear -->
            <div class="card-actions justify-end mt-3">
              <button class="btn btn-ghost btn-sm" onclick={() => { jobLogs = []; resetStatus(); }}>
                🧹 Clear Log
              </button>
            </div>
          </div>
        </div>
      {/if}
    </div>
  </main>
</div>
