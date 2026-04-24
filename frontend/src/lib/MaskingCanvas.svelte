<script>
  // MaskingCanvas.svelte
  // A canvas overlay that lets users click-and-drag to draw ignore-region masks
  // on top of a baseline screenshot image.

  const { src = '', targetUrl = '', existingMasks = [], onMaskCreated = () => {}, onMaskDeleted = () => {} } = $props();

  let canvasEl = $state(null);
  let containerEl = $state(null);
  let imgEl = $state(null);
  let drawing = $state(false);
  let startX = $state(0);
  let startY = $state(0);
  let currentX = $state(0);
  let currentY = $state(0);
  let imgLoaded = $state(false);
  let naturalWidth = $state(0);
  let naturalHeight = $state(0);
  let displayWidth = $state(0);
  let displayHeight = $state(0);
  let labelInput = $state('');
  let showLabelDialog = $state(false);
  let pendingRect = $state(null);

  function onImageLoad(e) {
    naturalWidth = e.target.naturalWidth;
    naturalHeight = e.target.naturalHeight;
    imgLoaded = true;
    requestAnimationFrame(resizeCanvas);
  }

  function resizeCanvas() {
    if (!canvasEl || !imgEl) return;
    const rect = imgEl.getBoundingClientRect();
    displayWidth = rect.width;
    displayHeight = rect.height;
    canvasEl.width = rect.width;
    canvasEl.height = rect.height;
    redraw();
  }

  function getScaleFactors() {
    if (!naturalWidth || !displayWidth) return { sx: 1, sy: 1 };
    return {
      sx: naturalWidth / displayWidth,
      sy: naturalHeight / displayHeight,
    };
  }

  function onMouseDown(e) {
    if (showLabelDialog) return;
    const rect = canvasEl.getBoundingClientRect();
    startX = e.clientX - rect.left;
    startY = e.clientY - rect.top;
    drawing = true;
  }

  function onMouseMove(e) {
    if (!drawing) return;
    const rect = canvasEl.getBoundingClientRect();
    currentX = e.clientX - rect.left;
    currentY = e.clientY - rect.top;
    redraw();
    // Draw the in-progress rectangle
    const ctx = canvasEl.getContext('2d');
    const x = Math.min(startX, currentX);
    const y = Math.min(startY, currentY);
    const w = Math.abs(currentX - startX);
    const h = Math.abs(currentY - startY);
    ctx.strokeStyle = '#00ff88';
    ctx.lineWidth = 2;
    ctx.setLineDash([6, 3]);
    ctx.strokeRect(x, y, w, h);
    ctx.fillStyle = 'rgba(0, 255, 136, 0.15)';
    ctx.fillRect(x, y, w, h);
    ctx.setLineDash([]);
  }

  function onMouseUp() {
    if (!drawing) return;
    drawing = false;
    const { sx, sy } = getScaleFactors();
    const x = Math.min(startX, currentX);
    const y = Math.min(startY, currentY);
    const w = Math.abs(currentX - startX);
    const h = Math.abs(currentY - startY);

    // Ignore tiny accidental drags
    if (w < 5 || h < 5) return;

    // Scale to original image coordinates
    pendingRect = {
      x: Math.round(x * sx),
      y: Math.round(y * sy),
      width: Math.round(w * sx),
      height: Math.round(h * sy),
    };
    labelInput = '';
    showLabelDialog = true;
  }

  function confirmMask() {
    if (!pendingRect) return;
    onMaskCreated({
      target_url: targetUrl,
      x: pendingRect.x,
      y: pendingRect.y,
      width: pendingRect.width,
      height: pendingRect.height,
      label: labelInput || 'Untitled Mask',
    });
    showLabelDialog = false;
    pendingRect = null;
    redraw();
  }

  function cancelMask() {
    showLabelDialog = false;
    pendingRect = null;
    redraw();
  }

  function redraw() {
    if (!canvasEl) return;
    const ctx = canvasEl.getContext('2d');
    ctx.clearRect(0, 0, canvasEl.width, canvasEl.height);

    const { sx, sy } = getScaleFactors();

    // Draw existing masks
    for (const mask of existingMasks) {
      const dx = mask.x / sx;
      const dy = mask.y / sy;
      const dw = mask.width / sx;
      const dh = mask.height / sy;

      ctx.fillStyle = 'rgba(255, 59, 48, 0.25)';
      ctx.fillRect(dx, dy, dw, dh);
      ctx.strokeStyle = '#ff3b30';
      ctx.lineWidth = 2;
      ctx.strokeRect(dx, dy, dw, dh);

      // Label
      ctx.font = '11px Inter, system-ui, sans-serif';
      ctx.fillStyle = '#ff3b30';
      const labelText = mask.label || `Mask #${mask.id}`;
      const textMetrics = ctx.measureText(labelText);
      ctx.fillStyle = 'rgba(0,0,0,0.7)';
      ctx.fillRect(dx, dy - 16, textMetrics.width + 8, 16);
      ctx.fillStyle = '#ff6b6b';
      ctx.fillText(labelText, dx + 4, dy - 4);

      // Delete button (small X in top-right corner)
      const btnSize = 18;
      const btnX = dx + dw - btnSize - 2;
      const btnY = dy + 2;
      ctx.fillStyle = 'rgba(255, 59, 48, 0.85)';
      ctx.fillRect(btnX, btnY, btnSize, btnSize);
      ctx.fillStyle = '#fff';
      ctx.font = 'bold 12px sans-serif';
      ctx.fillText('✕', btnX + 4, btnY + 13);
    }
  }

  function onCanvasClick(e) {
    if (drawing || showLabelDialog) return;
    const rect = canvasEl.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const clickY = e.clientY - rect.top;
    const { sx, sy } = getScaleFactors();

    // Check if click is on any mask's delete button
    for (const mask of existingMasks) {
      const dx = mask.x / sx;
      const dy = mask.y / sy;
      const dw = mask.width / sx;
      const btnSize = 18;
      const btnX = dx + dw - btnSize - 2;
      const btnY = dy + 2;
      if (clickX >= btnX && clickX <= btnX + btnSize && clickY >= btnY && clickY <= btnY + btnSize) {
        onMaskDeleted(mask.id);
        return;
      }
    }
  }

  $effect(() => {
    if (imgLoaded && canvasEl) {
      resizeCanvas();
    }
  });

  $effect(() => {
    // Re-render whenever existingMasks changes
    if (existingMasks && canvasEl) {
      redraw();
    }
  });
</script>

<div class="masking-canvas-wrapper" bind:this={containerEl}>
  <div class="canvas-container">
    <img
      bind:this={imgEl}
      {src}
      alt="Baseline Screenshot"
      class="baseline-img"
      onload={onImageLoad}
      draggable="false"
    />
    {#if imgLoaded}
      <canvas
        bind:this={canvasEl}
        class="mask-overlay"
        onmousedown={onMouseDown}
        onmousemove={onMouseMove}
        onmouseup={onMouseUp}
        onmouseleave={onMouseUp}
        onclick={onCanvasClick}
      ></canvas>
    {/if}
  </div>

  {#if showLabelDialog}
    <div class="label-dialog-overlay">
      <div class="label-dialog">
        <h4>🏷️ Name this Mask Region</h4>
        <p class="dialog-hint">
          Region: {pendingRect?.width}×{pendingRect?.height}px at ({pendingRect?.x}, {pendingRect?.y})
        </p>
        <input
          type="text"
          class="input input-bordered input-sm w-full"
          placeholder="e.g. Ad Banner, Chat Widget, Timestamp..."
          bind:value={labelInput}
          onkeydown={(e) => e.key === 'Enter' && confirmMask()}
          autofocus
        />
        <div class="dialog-actions">
          <button class="btn btn-ghost btn-sm" onclick={cancelMask}>Cancel</button>
          <button class="btn btn-primary btn-sm" onclick={confirmMask}>✅ Save Mask</button>
        </div>
      </div>
    </div>
  {/if}

  <div class="canvas-instructions">
    <span class="instruction-icon">🖱️</span>
    <span>Click and drag on the image to draw an ignore region. Masked areas will be excluded from visual comparison.</span>
  </div>
</div>

<style>
  .masking-canvas-wrapper {
    position: relative;
    width: 100%;
  }

  .canvas-container {
    position: relative;
    display: inline-block;
    width: 100%;
    border-radius: 12px;
    overflow: hidden;
    border: 2px solid hsl(var(--p) / 0.3);
    background: #0a0a0a;
  }

  .baseline-img {
    display: block;
    width: 100%;
    height: auto;
    user-select: none;
    pointer-events: none;
  }

  .mask-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    cursor: crosshair;
    z-index: 10;
  }

  .canvas-instructions {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
    padding: 8px 12px;
    background: hsl(var(--b2));
    border-radius: 8px;
    font-size: 12px;
    color: hsl(var(--bc) / 0.5);
  }

  .instruction-icon {
    font-size: 16px;
    flex-shrink: 0;
  }

  .label-dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    backdrop-filter: blur(4px);
  }

  .label-dialog {
    background: hsl(var(--b2));
    border: 1px solid hsl(var(--bc) / 0.1);
    border-radius: 16px;
    padding: 24px;
    min-width: 360px;
    box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  }

  .label-dialog h4 {
    margin: 0 0 4px 0;
    font-size: 16px;
    font-weight: 700;
  }

  .dialog-hint {
    margin: 0 0 12px 0;
    font-size: 11px;
    opacity: 0.5;
    font-family: 'JetBrains Mono', monospace;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 12px;
  }
</style>
