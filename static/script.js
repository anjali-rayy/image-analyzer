// ── State ──
    let mode = 'single';
    let singleFile = null;
    let batchFiles = [];
    let lastSingleResult = null;
    let lastBatchResults = null;

    const CIRC = 2 * Math.PI * 34; // ~213.6

    // ── Mode switch ──
    function switchMode(m) {
      mode = m;
      document.getElementById('modeSingle').classList.toggle('active', m === 'single');
      document.getElementById('modeBatch').classList.toggle('active', m === 'batch');
      document.getElementById('fileInput').multiple = (m === 'batch');
      document.getElementById('uploadHint').textContent =
        m === 'batch' ? 'JPG, PNG · up to 20 files' : 'JPG, PNG · max 10 MB per file';
      resetAll();
    }

    // ── Upload ──
    const uploadArea = document.getElementById('uploadArea');
    const fileInput  = document.getElementById('fileInput');

    uploadArea.addEventListener('click', () => fileInput.click());
    uploadArea.addEventListener('dragover', e => { e.preventDefault(); uploadArea.classList.add('dragover'); });
    uploadArea.addEventListener('dragleave', () => uploadArea.classList.remove('dragover'));
    uploadArea.addEventListener('drop', e => {
      e.preventDefault();
      uploadArea.classList.remove('dragover');
      handleFiles(e.dataTransfer.files);
    });
    fileInput.addEventListener('change', () => handleFiles(fileInput.files));

    function handleFiles(files) {
      if (!files || files.length === 0) return;
      mode === 'single' ? handleSingleFile(files[0]) : handleBatchFiles(files);
    }

    // ── Single ──
    function handleSingleFile(file) {
      singleFile = file;
      document.getElementById('previewThumb').src = URL.createObjectURL(file);
      document.getElementById('previewName').textContent = file.name;
      document.getElementById('previewMeta').textContent =
        formatSize(file.size) + ' · ' + file.type.replace('image/', '').toUpperCase();
      showOnly('preview-section');
    }

    document.getElementById('analyzeBtn').addEventListener('click', async () => {
      if (!singleFile) return;
      const btn = document.getElementById('analyzeBtn');
      btn.disabled = true;
      setLoading(true, 'Analyzing your image…');

      const form = new FormData();
      form.append('image', singleFile);

      try {
        const res  = await fetch('/analyze', { method:'POST', body:form });
        const data = await res.json();
        lastSingleResult = data;
        showSingleResult(data);
      } catch (e) {
        alert('Error connecting to server. Make sure it is running.');
        setLoading(false);
        showOnly('preview-section');
      } finally {
        btn.disabled = false;
      }
    });

    function showSingleResult(data) {
      setLoading(false);
      showOnly('results');

      const score = Math.round(data.overall_score);

      // Arc animation
      const arc = document.getElementById('scoreArc');
      const offset = CIRC - (score / 100) * CIRC;
      let color = '#c46a6a';
      if (score >= 70) color = '#5fa86e';
      else if (score >= 40) color = '#c4a44a';

      setTimeout(() => {
        arc.style.strokeDashoffset = offset;
        arc.style.stroke = color;
      }, 50);

      document.getElementById('scoreNum').textContent = score;
      document.getElementById('scoreNum').style.color = color;

      let title = 'Needs Work', sub = 'Several quality issues detected in this photo.';
      if (score >= 70) { title = 'Great Shot';   sub = 'Technical quality looks solid. Good to use.'; }
      else if (score >= 40) { title = 'Acceptable'; sub = 'Some quality issues worth noting.'; }

      document.getElementById('scoreTitle').textContent    = title;
      document.getElementById('scoreSubtitle').textContent = sub;

      // Info
      document.getElementById('rDimensions').textContent = data.width + ' × ' + data.height + ' px';
      document.getElementById('rAspect').textContent     = data.aspect_ratio;
      document.getElementById('rFileSize').textContent   = formatSize(data.file_size);

      // Sharpness bar (0–20 range, cap at 20)
      const sharpPct = Math.min(100, (data.blur.sharpness / 20) * 100);
      setBar('rSharpBar', sharpPct, data.blur.status === 'sharp' ? 'var(--good)' : 'var(--bad)');
      document.getElementById('rSharpness').textContent = data.blur.sharpness;
      setBadge('rBlurBadge', data.blur.status, { sharp:'good', blurry:'bad' });

      // Brightness bar (0–100)
      setBar('rBrightBar', data.brightness.value, 'var(--warn)');
      document.getElementById('rBrightness').textContent = data.brightness.value;
      setBadge('rBrightBadge', data.brightness.status, {
        good:'good', 'too dark':'bad', 'too bright':'warn'
      });

      // Noise bar (0–10, cap at 10, INVERTED — lower is better)
      const noisePct = Math.min(100, (data.noise.level / 10) * 100);
      setBar('rNoiseBar', noisePct, data.noise.status === 'clean' ? 'var(--good)' : data.noise.status === 'noisy' ? 'var(--bad)' : 'var(--warn)');
      document.getElementById('rNoise').textContent = data.noise.level;
      setBadge('rNoiseBadge', data.noise.status, { clean:'good', moderate:'warn', noisy:'bad' });

      // Color Profile
      const cp = data.color_profile;
      document.getElementById('rDominant').textContent = cp.dominant || 'N/A';
      setBar('rVibranceBar', cp.vibrance, 'var(--accent)');
      document.getElementById('rVibrance').textContent = cp.vibrance;
      setBadge('rColorfulBadge', cp.colorful ? 'colorful' : 'muted', { colorful:'good', muted:'neutral' });
      document.getElementById('rAvgRGB').textContent =
        'R:' + Math.round(cp.avg_r) + ' G:' + Math.round(cp.avg_g) + ' B:' + Math.round(cp.avg_b);

      // EXIF
      document.getElementById('rCamera').textContent   = data.exif.camera        || 'N/A';
      document.getElementById('rISO').textContent      = data.exif.iso           || 'N/A';
      document.getElementById('rFocal').textContent    = data.exif.focal_length  || 'N/A';
      document.getElementById('rExposure').textContent = data.exif.exposure_time || 'N/A';
    }

    // ── Batch ──
    function handleBatchFiles(files) {
      const arr = Array.from(files).slice(0, 20);
      batchFiles = [...batchFiles, ...arr].slice(0, 20);
      renderBatchList();
      showOnly('batch-section');
    }

    function renderBatchList() {
      const list = document.getElementById('batchList');
      document.getElementById('batchCount').textContent =
        batchFiles.length + ' image' + (batchFiles.length !== 1 ? 's' : '');
      list.innerHTML = '';
      batchFiles.forEach((f, i) => {
        const url = URL.createObjectURL(f);
        const div = document.createElement('div');
        div.className = 'batch-item';
        div.innerHTML = `
          <img class="batch-thumb" src="${url}" alt="">
          <div class="batch-item-info">
            <div class="batch-item-name">${escHtml(f.name)}</div>
            <div class="batch-item-size">${formatSize(f.size)}</div>
          </div>
          <button class="batch-item-remove" onclick="removeBatchItem(${i})" title="Remove">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>`;
        list.appendChild(div);
      });
    }

    function removeBatchItem(idx) {
      batchFiles.splice(idx, 1);
      if (batchFiles.length === 0) { resetAll(); return; }
      renderBatchList();
    }

    document.getElementById('batchAnalyzeBtn').addEventListener('click', async () => {
      if (batchFiles.length === 0) return;
      const btn = document.getElementById('batchAnalyzeBtn');
      btn.disabled = true;
      setLoading(true, `Analyzing ${batchFiles.length} image${batchFiles.length > 1 ? 's' : ''}…`);

      const form = new FormData();
      batchFiles.forEach(f => form.append('images', f));

      try {
        const res  = await fetch('/analyze-batch', { method:'POST', body:form });
        const data = await res.json();
        lastBatchResults = data.results;
        showBatchResults(data);
      } catch (e) {
        alert('Error connecting to server.');
        setLoading(false);
        showOnly('batch-section');
      } finally {
        btn.disabled = false;
      }
    });

    function showBatchResults(data) {
      setLoading(false);
      showOnly('batch-results');
      document.getElementById('batchResultsTitle').textContent =
        `${data.total} Image${data.total !== 1 ? 's' : ''} Analyzed`;

      const tbody = document.getElementById('batchTableBody');
      tbody.innerHTML = '';

      data.results.forEach(r => {
        const score = Math.round(r.overall_score);
        const cls   = score >= 70 ? 'good' : score >= 40 ? 'ok' : 'bad';
        const blurCls  = r.blur.status === 'sharp'   ? 'good' : 'bad';
        const brCls    = r.brightness.status === 'good' ? 'good' : r.brightness.status === 'too dark' ? 'bad' : 'warn';
        const noiseCls = r.noise.status === 'clean'  ? 'good' : r.noise.status === 'noisy' ? 'bad' : 'warn';
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td title="${escHtml(r.file_name)}">${escHtml(r.file_name)}</td>
          <td>${formatSize(r.file_size)}</td>
          <td>${r.width}×${r.height}</td>
          <td>${r.blur.sharpness}</td>
          <td><span class="badge badge-${blurCls}">${r.blur.status}</span></td>
          <td><span class="badge badge-${brCls}">${r.brightness.status}</span></td>
          <td><span class="badge badge-${noiseCls}">${r.noise.status}</span></td>
          <td class="score-cell ${cls}">${score}</td>`;
        tbody.appendChild(tr);
      });
    }

    // ── CSV export ──
    document.getElementById('exportSingleCSV').addEventListener('click', () => {
      if (lastSingleResult) exportCSV([lastSingleResult]);
    });
    document.getElementById('exportBatchCSV').addEventListener('click', () => {
      if (lastBatchResults) exportCSV(lastBatchResults);
    });

    async function exportCSV(results) {
      try {
        const res = await fetch('/export-csv', {
          method:'POST',
          headers:{ 'Content-Type':'application/json' },
          body: JSON.stringify(results)
        });
        const blob = await res.blob();
        const url  = URL.createObjectURL(blob);
        const a    = document.createElement('a');
        a.href = url; a.download = 'image-analysis.csv'; a.click();
        URL.revokeObjectURL(url);
      } catch (e) {
        alert('Could not export CSV.');
      }
    }

    // ── Helpers ──
    function showOnly(id) {
      ['uploadWrap','preview-section','batch-section','loading','results','batch-results']
        .forEach(s => document.getElementById(s).style.display = 'none');
      document.getElementById(id).style.display = 'block';
    }

    function setLoading(on, text) {
      if (on) {
        document.getElementById('loadingText').textContent = text || 'Analyzing…';
        showOnly('loading');
      }
    }

    function resetAll() {
      singleFile = null; batchFiles = [];
      lastSingleResult = null; lastBatchResults = null;
      fileInput.value = '';
      // Reset arc
      const arc = document.getElementById('scoreArc');
      arc.style.strokeDashoffset = CIRC;
      ['preview-section','batch-section','loading','results','batch-results']
        .forEach(s => document.getElementById(s).style.display = 'none');
      document.getElementById('uploadWrap').style.display = 'block';
    }

    function setBar(id, pct, color) {
      const el = document.getElementById(id);
      el.style.background = color;
      setTimeout(() => { el.style.width = Math.min(100, Math.max(0, pct)) + '%'; }, 60);
    }

    function setBadge(id, text, map) {
      const el = document.getElementById(id);
      el.textContent = text;
      el.className   = 'badge badge-' + (map[text] || 'neutral');
    }

    function formatSize(bytes) {
      if (!bytes) return 'N/A';
      if (bytes < 1024) return bytes + ' B';
      if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
      return (bytes / 1048576).toFixed(2) + ' MB';
    }

    function escHtml(str) {
      return String(str)
        .replace(/&/g,'&amp;').replace(/</g,'&lt;')
        .replace(/>/g,'&gt;').replace(/"/g,'&quot;');
    }