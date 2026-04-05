// ═══════════════════════════════════════
// BITSBURG — 16-Bit Sprite & Tile System
// ═══════════════════════════════════════
// Pixel-perfect 16-bit JRPG style tiles,
// buildings, and characters matching the
// Bitsburg design reference images.
// ═══════════════════════════════════════

// ─── PALETTE ───
export const P = {
  // Grass tones (4-tone for depth)
  g1: '#4CAF50', g2: '#388E3C', g3: '#66BB6A', g4: '#2E7D32',
  // Dirt/path
  d1: '#C4A882', d2: '#B09872', d3: '#D4B892', d4: '#9E8862',
  // Stone
  s1: '#9E9E9E', s2: '#757575', s3: '#BDBDBD', s4: '#616161',
  // Water (animated)
  w1: '#42A5F5', w2: '#1E88E5', w3: '#64B5F6', w4: '#1565C0',
  // Wood
  wd1: '#8D6E63', wd2: '#6D4C41', wd3: '#A1887F', wd4: '#5D4037',
  // Roof colors
  r_red1: '#E53935', r_red2: '#C62828', r_red3: '#EF5350',
  r_blue1: '#1E88E5', r_blue2: '#1565C0', r_blue3: '#42A5F5',
  r_green1: '#43A047', r_green2: '#2E7D32', r_green3: '#66BB6A',
  r_brown1: '#6D4C41', r_brown2: '#5D4037', r_brown3: '#8D6E63',
  // Walls
  wl_white: '#FAFAFA', wl_brick: '#D7CCC8', wl_stone: '#BDBDBD',
  wl_wood: '#A1887F', wl_castle: '#9E9E9E',
  // Windows & doors
  win: '#BBDEFB', winFrame: '#5D4037', door: '#6D4C41', doorHandle: '#FFB300',
  // Character skins
  sk1: '#FFCCB3', sk2: '#E8B87E', sk3: '#C68642', sk4: '#8D5524',
  // Hair
  hr_black: '#212121', hr_brown: '#5D4037', hr_blonde: '#F9A825',
  hr_red: '#E53935', hr_blue: '#1E88E5', hr_green: '#43A047',
  hr_purple: '#8E24AA', hr_silver: '#BDBDBD',
  // Clothes
  cl_blue: '#1565C0', cl_red: '#C62828', cl_green: '#2E7D32',
  cl_purple: '#6A1B9A', cl_yellow: '#F9A825', cl_grey: '#616161',
  cl_black: '#212121', cl_white: '#FAFAFA', cl_orange: '#EF6C00',
  cl_pink: '#AD1457',
  // Accents
  gold: '#FFD54F', cyan: '#00E5FF', magenta: '#E040FB',
  black: '#1A1A2E', shadow: 'rgba(0,0,0,0.25)',
  tree_green1: '#2E7D32', tree_green2: '#388E3C', tree_green3: '#4CAF50',
  flower_red: '#E53935', flower_yellow: '#FDD835', flower_purple: '#8E24AA',
  fence: '#A1887F', fountain: '#42A5F5',
}

// ─── TILE TYPES ───
export const T = {
  GRASS: 0, GRASS_DARK: 1, PATH: 2, PATH_H: 3,
  WATER: 4, WATER_DEEP: 5, WALL: 6, WALL_DARK: 7,
  FLOOR: 8, BRIDGE: 9, SAND: 10, FLOWER: 11,
  TREE: 12, FENCE: 13, FOUNTAIN: 14, CROP: 15,
}

const TS = 16 // tile size in pixels

// ─── TILE RENDERER ───
// Each tile is a 16×16 pixel art tile with proper detail
const tileCache = {}

function getTileCanvas(type) {
  if (tileCache[type]) return tileCache[type]
  const c = document.createElement('canvas')
  c.width = TS; c.height = TS
  const ctx = c.getContext('2d')
  drawTilePixel(ctx, type)
  tileCache[type] = c
  return c
}

function drawTilePixel(ctx, type) {
  ctx.clearRect(0, 0, TS, TS)
  
  switch (type) {
    case T.GRASS:
      // Base green with texture
      ctx.fillStyle = P.g1
      ctx.fillRect(0, 0, TS, TS)
      // Grass blades
      ctx.fillStyle = P.g2
      ctx.fillRect(2, 3, 1, 4); ctx.fillRect(7, 8, 1, 3)
      ctx.fillRect(12, 2, 1, 5); ctx.fillRect(5, 12, 1, 3)
      ctx.fillRect(14, 10, 1, 4)
      // Highlights
      ctx.fillStyle = P.g3
      ctx.fillRect(3, 1, 1, 2); ctx.fillRect(9, 6, 1, 2)
      ctx.fillRect(1, 9, 1, 2); ctx.fillRect(11, 13, 1, 2)
      break
      
    case T.GRASS_DARK:
      ctx.fillStyle = P.g2
      ctx.fillRect(0, 0, TS, TS)
      ctx.fillStyle = P.g4
      ctx.fillRect(3, 4, 1, 3); ctx.fillRect(8, 1, 1, 4)
      ctx.fillRect(13, 7, 1, 3); ctx.fillRect(1, 11, 1, 4)
      ctx.fillRect(10, 13, 1, 2)
      ctx.fillStyle = P.g1
      ctx.fillRect(5, 2, 1, 2); ctx.fillRect(11, 5, 1, 2)
      break
      
    case T.PATH:
      // Stone path with texture
      ctx.fillStyle = P.d1
      ctx.fillRect(0, 0, TS, TS)
      // Stone pattern
      ctx.fillStyle = P.d2
      ctx.fillRect(0, 0, TS, 1); ctx.fillRect(0, TS-1, TS, 1)
      ctx.fillRect(0, 0, 1, TS); ctx.fillRect(TS-1, 0, 1, TS)
      ctx.fillStyle = P.d3
      ctx.fillRect(2, 2, 5, 5); ctx.fillRect(9, 9, 5, 5)
      ctx.fillStyle = P.d4
      ctx.fillRect(8, 2, 1, 1); ctx.fillRect(3, 8, 1, 1)
      ctx.fillRect(12, 6, 1, 1); ctx.fillRect(5, 13, 1, 1)
      break
      
    case T.PATH_H:
      // Horizontal path variant
      ctx.fillStyle = P.d1
      ctx.fillRect(0, 0, TS, TS)
      ctx.fillStyle = P.d2
      ctx.fillRect(0, 0, TS, 1); ctx.fillRect(0, TS-1, TS, 1)
      ctx.fillStyle = P.d3
      ctx.fillRect(1, 4, 6, 4); ctx.fillRect(9, 8, 6, 4)
      ctx.fillStyle = P.d4
      ctx.fillRect(4, 2, 1, 1); ctx.fillRect(11, 12, 1, 1)
      break
      
    case T.WATER:
      // Animated water with waves
      const t = Date.now() / 800
      ctx.fillStyle = P.w2
      ctx.fillRect(0, 0, TS, TS)
      // Wave lines
      ctx.fillStyle = P.w1
      for (let i = 0; i < 3; i++) {
        const wy = 3 + i * 5 + Math.sin(t + i) * 1
        ctx.fillRect(1, Math.floor(wy), 4, 1)
        ctx.fillRect(7, Math.floor(wy + 1), 4, 1)
      }
      // Shimmer
      ctx.fillStyle = P.w3
      ctx.fillRect(3, 1, 2, 1); ctx.fillRect(10, 8, 2, 1)
      ctx.fillRect(6, 13, 2, 1)
      break
      
    case T.WATER_DEEP:
      ctx.fillStyle = P.w4
      ctx.fillRect(0, 0, TS, TS)
      ctx.fillStyle = P.w2
      const t2 = Date.now() / 600
      for (let i = 0; i < 4; i++) {
        const wy = 2 + i * 4 + Math.sin(t2 + i * 0.7) * 1
        ctx.fillRect(2, Math.floor(wy), 3, 1)
        ctx.fillRect(9, Math.floor(wy + 0.5), 3, 1)
      }
      ctx.fillStyle = P.w3
      ctx.fillRect(5, 3, 1, 1); ctx.fillRect(12, 10, 1, 1)
      break
      
    case T.WALL:
      // Stone wall
      ctx.fillStyle = P.s1
      ctx.fillRect(0, 0, TS, TS)
      // Brick pattern
      ctx.fillStyle = P.s2
      ctx.fillRect(0, 0, TS, 1); ctx.fillRect(0, 7, TS, 1)
      ctx.fillRect(0, 15, TS, 1)
      ctx.fillRect(0, 0, 1, TS); ctx.fillRect(7, 0, 1, 8)
      ctx.fillRect(7, 8, 1, 8); ctx.fillRect(15, 0, 1, TS)
      // Mortar lines
      ctx.fillStyle = P.s3
      ctx.fillRect(1, 1, 6, 1); ctx.fillRect(8, 8, 7, 1)
      ctx.fillRect(1, 8, 6, 1)
      // Highlight
      ctx.fillStyle = '#E0E0E0'
      ctx.fillRect(1, 2, 1, 1); ctx.fillRect(8, 9, 1, 1)
      break
      
    case T.WALL_DARK:
      ctx.fillStyle = P.s2
      ctx.fillRect(0, 0, TS, TS)
      ctx.fillStyle = P.s4
      ctx.fillRect(0, 0, TS, 1); ctx.fillRect(0, 7, TS, 1)
      ctx.fillRect(0, 15, TS, 1)
      ctx.fillRect(0, 0, 1, TS); ctx.fillRect(7, 0, 1, 8)
      ctx.fillRect(7, 8, 1, 8); ctx.fillRect(15, 0, 1, TS)
      ctx.fillStyle = P.s1
      ctx.fillRect(1, 1, 6, 1); ctx.fillRect(8, 8, 7, 1)
      break
      
    case T.FLOOR:
      // Wooden floor
      ctx.fillStyle = P.wd1
      ctx.fillRect(0, 0, TS, TS)
      // Plank lines
      ctx.fillStyle = P.wd2
      ctx.fillRect(0, 3, TS, 1); ctx.fillRect(0, 7, TS, 1)
      ctx.fillRect(0, 11, TS, 1); ctx.fillRect(0, 15, TS, 1)
      // Wood grain
      ctx.fillStyle = P.wd3
      ctx.fillRect(2, 1, 1, 2); ctx.fillRect(10, 5, 1, 2)
      ctx.fillRect(5, 9, 1, 2); ctx.fillRect(13, 13, 1, 2)
      break
      
    case T.BRIDGE:
      // Wooden bridge over water
      ctx.fillStyle = P.w2
      ctx.fillRect(0, 0, TS, TS)
      ctx.fillStyle = P.wd1
      ctx.fillRect(1, 0, 14, TS)
      // Planks
      ctx.fillStyle = P.wd2
      ctx.fillRect(1, 0, 14, 2); ctx.fillRect(1, 4, 14, 1)
      ctx.fillRect(1, 8, 14, 1); ctx.fillRect(1, 12, 14, 1)
      ctx.fillRect(1, 14, 14, 2)
      // Railings
      ctx.fillStyle = P.wd3
      ctx.fillRect(1, 1, 1, TS - 2)
      ctx.fillRect(14, 1, 1, TS - 2)
      break
      
    case T.SAND:
      ctx.fillStyle = P.d3
      ctx.fillRect(0, 0, TS, TS)
      ctx.fillStyle = P.d2
      ctx.fillRect(2, 3, 2, 1); ctx.fillRect(8, 7, 3, 1)
      ctx.fillRect(4, 12, 2, 1); ctx.fillRect(11, 2, 1, 1)
      ctx.fillRect(1, 9, 2, 1)
      ctx.fillStyle = P.d4
      ctx.fillRect(6, 1, 1, 1); ctx.fillRect(13, 10, 1, 1)
      break
      
    case T.FLOWER:
      ctx.fillStyle = P.g1
      ctx.fillRect(0, 0, TS, TS)
      ctx.fillStyle = P.g2
      ctx.fillRect(2, 3, 1, 4); ctx.fillRect(7, 8, 1, 3)
      ctx.fillRect(12, 2, 1, 5)
      // Flowers
      ctx.fillStyle = P.flower_red
      ctx.fillRect(3, 5, 2, 2); ctx.fillRect(9, 10, 2, 2)
      ctx.fillStyle = P.flower_yellow
      ctx.fillRect(11, 4, 2, 2)
      ctx.fillStyle = P.flower_purple
      ctx.fillRect(1, 11, 2, 2)
      // Centers
      ctx.fillStyle = '#FFF176'
      ctx.fillRect(4, 6, 1, 1); ctx.fillRect(10, 11, 1, 1)
      ctx.fillRect(12, 5, 1, 1); ctx.fillRect(2, 12, 1, 1)
      break
      
    case T.TREE:
      ctx.fillStyle = P.g1
      ctx.fillRect(0, 0, TS, TS)
      // Trunk
      ctx.fillStyle = P.wd2
      ctx.fillRect(6, 10, 4, 6)
      ctx.fillStyle = P.wd3
      ctx.fillRect(7, 10, 1, 6)
      // Canopy
      ctx.fillStyle = P.tree_green2
      ctx.fillRect(2, 2, 12, 10)
      ctx.fillStyle = P.tree_green1
      ctx.fillRect(3, 3, 10, 3); ctx.fillRect(2, 6, 3, 4)
      ctx.fillRect(11, 6, 3, 4)
      ctx.fillStyle = P.tree_green3
      ctx.fillRect(4, 1, 8, 2); ctx.fillRect(5, 4, 6, 2)
      ctx.fillRect(3, 7, 2, 2); ctx.fillRect(11, 7, 2, 2)
      // Highlight
      ctx.fillStyle = '#81C784'
      ctx.fillRect(5, 2, 3, 1); ctx.fillRect(8, 5, 2, 1)
      break
      
    case T.FENCE:
      ctx.fillStyle = P.g1
      ctx.fillRect(0, 0, TS, TS)
      // Posts
      ctx.fillStyle = P.fence
      ctx.fillRect(1, 2, 2, 14); ctx.fillRect(7, 2, 2, 14)
      ctx.fillRect(13, 2, 2, 14)
      // Rails
      ctx.fillStyle = P.wd3
      ctx.fillRect(0, 5, TS, 2); ctx.fillRect(0, 11, TS, 2)
      // Post tops
      ctx.fillStyle = P.wd2
      ctx.fillRect(1, 1, 2, 2); ctx.fillRect(7, 1, 2, 2)
      ctx.fillRect(13, 1, 2, 2)
      break
      
    case T.FOUNTAIN:
      ctx.fillStyle = P.d1
      ctx.fillRect(0, 0, TS, TS)
      // Base
      ctx.fillStyle = P.s1
      ctx.fillRect(2, 2, 12, 12)
      ctx.fillStyle = P.s2
      ctx.fillRect(3, 3, 10, 10)
      // Water
      ctx.fillStyle = P.fountain
      ctx.fillRect(4, 4, 8, 8)
      ctx.fillStyle = P.w3
      const ft = Date.now() / 500
      ctx.fillRect(5 + Math.sin(ft) * 2, 5, 2, 1)
      ctx.fillRect(8, 8 + Math.cos(ft) * 1, 2, 1)
      // Edge stones
      ctx.fillStyle = P.s3
      ctx.fillRect(2, 2, 12, 1); ctx.fillRect(2, 13, 12, 1)
      ctx.fillRect(2, 2, 1, 12); ctx.fillRect(13, 2, 1, 12)
      break
      
    case T.CROP:
      ctx.fillStyle = P.d1
      ctx.fillRect(0, 0, TS, TS)
      // Soil rows
      ctx.fillStyle = P.wd4
      ctx.fillRect(0, 3, TS, 2); ctx.fillRect(0, 9, TS, 2)
      // Plants
      ctx.fillStyle = P.g3
      ctx.fillRect(2, 1, 1, 3); ctx.fillRect(5, 1, 1, 3)
      ctx.fillRect(9, 1, 1, 3); ctx.fillRect(13, 1, 1, 3)
      ctx.fillRect(2, 7, 1, 3); ctx.fillRect(6, 7, 1, 3)
      ctx.fillRect(10, 7, 1, 3); ctx.fillRect(14, 7, 1, 3)
      // Wheat tips
      ctx.fillStyle = P.gold
      ctx.fillRect(2, 0, 1, 1); ctx.fillRect(5, 0, 1, 1)
      ctx.fillRect(9, 0, 1, 1); ctx.fillRect(13, 0, 1, 1)
      ctx.fillRect(2, 6, 1, 1); ctx.fillRect(6, 6, 1, 1)
      ctx.fillRect(10, 6, 1, 1); ctx.fillRect(14, 6, 1, 1)
      break
      
    default:
      ctx.fillStyle = P.g1
      ctx.fillRect(0, 0, TS, TS)
  }
}

// ─── CHARACTER RENDERER ───
// 16×24 pixel characters with detailed body parts
// 4 directions × 2 walk frames = 8 sprite states
const charCache = {}

function getCharCanvas(config, dir, frame) {
  const key = `${config.skinColor}_${config.hairColor}_${config.clothColor}_${config.hairStyle}_${dir}_${frame}`
  if (charCache[key]) return charCache[key]
  const c = document.createElement('canvas')
  c.width = 16; c.height = 24
  const ctx = c.getContext('2d')
  drawCharPixel(ctx, config, dir, frame)
  charCache[key] = c
  return c
}

function drawCharPixel(ctx, cfg, dir, frame) {
  ctx.clearRect(0, 0, 16, 24)
  const { skinColor, hairColor, clothColor, hairStyle = 'short' } = cfg
  
  const walkOffset = frame === 1 ? 1 : 0
  
  // ── SHADOW ──
  ctx.fillStyle = P.shadow
  ctx.fillRect(3, 22, 10, 2)
  
  // ── FEET / LEGS ──
  const legColor = '#37474F'
  const bootColor = '#5D4037'
  
  if (dir === 0) { // Facing DOWN
    // Left leg
    ctx.fillStyle = legColor
    ctx.fillRect(5, 18, 3, 3 + walkOffset)
    ctx.fillStyle = bootColor
    ctx.fillRect(5, 21 + walkOffset, 3, 1)
    // Right leg
    ctx.fillStyle = legColor
    ctx.fillRect(8, 18, 3, 3 - walkOffset)
    ctx.fillStyle = bootColor
    ctx.fillRect(8, 21 - walkOffset, 3, 1)
  } else if (dir === 1) { // Facing UP
    ctx.fillStyle = legColor
    ctx.fillRect(5, 18, 3, 4)
    ctx.fillStyle = bootColor
    ctx.fillRect(5, 22, 3, 1)
    ctx.fillStyle = legColor
    ctx.fillRect(8, 18, 3, 4)
    ctx.fillStyle = bootColor
    ctx.fillRect(8, 22, 3, 1)
  } else if (dir === 2) { // Facing LEFT
    ctx.fillStyle = legColor
    ctx.fillRect(4, 18, 3, 3 + walkOffset)
    ctx.fillStyle = bootColor
    ctx.fillRect(4, 21 + walkOffset, 3, 1)
    ctx.fillStyle = legColor
    ctx.fillRect(8, 18, 3, 3 - walkOffset)
    ctx.fillStyle = bootColor
    ctx.fillRect(8, 21 - walkOffset, 3, 1)
  } else { // Facing RIGHT
    ctx.fillStyle = legColor
    ctx.fillRect(5, 18, 3, 3 + walkOffset)
    ctx.fillStyle = bootColor
    ctx.fillRect(5, 21 + walkOffset, 3, 1)
    ctx.fillStyle = legColor
    ctx.fillRect(9, 18, 3, 3 - walkOffset)
    ctx.fillStyle = bootColor
    ctx.fillRect(9, 21 - walkOffset, 3, 1)
  }
  
  // ── BODY / TORSO ──
  ctx.fillStyle = clothColor
  if (dir === 0) {
    // Front view
    ctx.fillRect(4, 11, 8, 7)
    // Collar detail
    ctx.fillStyle = shadeColor(clothColor, 30)
    ctx.fillRect(5, 11, 6, 1)
    // Button/detail
    ctx.fillStyle = shadeColor(clothColor, -20)
    ctx.fillRect(7, 14, 2, 1)
  } else if (dir === 1) {
    // Back view
    ctx.fillRect(4, 11, 8, 7)
    ctx.fillStyle = shadeColor(clothColor, -15)
    ctx.fillRect(4, 11, 8, 2)
  } else {
    // Side view
    ctx.fillRect(4, 11, 7, 7)
    ctx.fillStyle = shadeColor(clothColor, 20)
    ctx.fillRect(4, 11, 1, 7)
    // Arm
    ctx.fillStyle = shadeColor(clothColor, -10)
    ctx.fillRect(dir === 2 ? 3 : 10, 12, 2, 5)
  }
  
  // ── ARMS ──
  if (dir === 0 || dir === 1) {
    // Front/back arms
    ctx.fillStyle = shadeColor(clothColor, -15)
    const armSwing = frame === 1 ? 1 : 0
    ctx.fillRect(3, 12, 2, 5 + armSwing)
    ctx.fillRect(11, 12 - armSwing, 2, 5)
    // Hands
    ctx.fillStyle = skinColor
    ctx.fillRect(3, 17 + armSwing, 2, 1)
    ctx.fillRect(11, 17 - armSwing, 2, 1)
  }
  
  // ── HEAD ─
  ctx.fillStyle = skinColor
  ctx.fillRect(4, 3, 8, 8)
  
  // ── HAIR ──
  ctx.fillStyle = hairColor
  if (hairStyle === 'short') {
    ctx.fillRect(4, 1, 8, 4)
    ctx.fillRect(3, 3, 2, 3)
    ctx.fillRect(11, 3, 2, 3)
    if (dir === 0) {
      // Bangs
      ctx.fillRect(4, 5, 8, 2)
    }
  } else if (hairStyle === 'long') {
    ctx.fillRect(3, 1, 10, 4)
    ctx.fillRect(3, 3, 2, 8)
    ctx.fillRect(11, 3, 2, 8)
    if (dir === 0) {
      ctx.fillRect(4, 5, 8, 2)
    }
    if (dir === 1) {
      // Ponytail
      ctx.fillRect(5, 9, 6, 3)
    }
  } else if (hairStyle === 'robot') {
    ctx.fillRect(4, 0, 8, 4)
    ctx.fillRect(5, 4, 2, 2)
    ctx.fillRect(9, 4, 2, 2)
    // Antenna
    ctx.fillStyle = '#8899AA'
    ctx.fillRect(7, 0, 2, 2)
    ctx.fillStyle = '#FF5252'
    ctx.fillRect(7, -1, 2, 1)
  } else if (hairStyle === 'cap') {
    ctx.fillRect(3, 2, 10, 3)
    ctx.fillRect(10, 3, 4, 2)
  } else if (hairStyle === 'ponytail') {
    ctx.fillRect(4, 1, 8, 4)
    ctx.fillRect(2, 3, 3, 5)
  } else if (hairStyle === 'slick') {
    ctx.fillRect(4, 1, 8, 3)
    ctx.fillRect(3, 3, 2, 2)
    ctx.fillRect(11, 3, 2, 2)
  } else if (hairStyle === 'hat') {
    ctx.fillRect(3, 2, 10, 2)
    ctx.fillRect(1, 3, 14, 2)
  }
  
  // ── FACE ──
  if (dir === 0) {
    // Eyes (front)
    ctx.fillStyle = '#FFFFFF'
    ctx.fillRect(5, 6, 2, 2)
    ctx.fillRect(9, 6, 2, 2)
    ctx.fillStyle = '#212121'
    ctx.fillRect(6, 7, 1, 1)
    ctx.fillRect(10, 7, 1, 1)
    // Mouth
    ctx.fillStyle = shadeColor(skinColor, -30)
    ctx.fillRect(7, 9, 2, 1)
  } else if (dir === 1) {
    // Back of head - no face
  } else if (dir === 2) {
    // Left profile
    ctx.fillStyle = '#FFFFFF'
    ctx.fillRect(6, 6, 2, 2)
    ctx.fillStyle = '#212121'
    ctx.fillRect(6, 7, 1, 1)
    // Nose
    ctx.fillStyle = shadeColor(skinColor, -20)
    ctx.fillRect(4, 8, 1, 1)
  } else {
    // Right profile
    ctx.fillStyle = '#FFFFFF'
    ctx.fillRect(8, 6, 2, 2)
    ctx.fillStyle = '#212121'
    ctx.fillRect(9, 7, 1, 1)
    ctx.fillStyle = shadeColor(skinColor, -20)
    ctx.fillRect(11, 8, 1, 1)
  }
}

// Helper: lighten/darken hex color
function shadeColor(hex, percent) {
  const num = parseInt(hex.replace('#', ''), 16)
  const r = Math.min(255, Math.max(0, (num >> 16) + percent))
  const g = Math.min(255, Math.max(0, ((num >> 8) & 0x00FF) + percent))
  const b = Math.min(255, Math.max(0, (num & 0x0000FF) + percent))
  return `rgb(${r},${g},${b})`
}

// ─── BUILDING RENDERER ───
// Each building is rendered as a multi-tile structure with proper pixel art
const buildCache = {}

function getBuildingCanvas(type, w, h) {
  const key = `${type}_${w}_${h}`
  if (buildCache[key]) return buildCache[key]
  const pw = w * TS
  const ph = h * TS
  const c = document.createElement('canvas')
  c.width = pw; c.height = ph
  const ctx = c.getContext('2d')
  drawBuildingPixel(ctx, type, w, h)
  buildCache[key] = c
  return c
}

function drawBuildingPixel(ctx, type, w, h) {
  const pw = w * TS
  const ph = h * TS
  
  ctx.clearRect(0, 0, pw, ph)
  
  // Shadow
  ctx.fillStyle = 'rgba(0,0,0,0.2)'
  ctx.fillRect(2, ph - 4, pw - 2, 6)
  
  // Foundation
  ctx.fillStyle = P.s2
  ctx.fillRect(0, ph - TS, pw, TS)
  ctx.fillStyle = P.s3
  ctx.fillRect(0, ph - TS, pw, 2)
  
  // Wall
  const wallTop = h > 2 ? TS : TS * 0.6
  if (type === 'castle') {
    // Castle walls (stone)
    for (let y = wallTop; y < ph - TS; y += TS) {
      for (let x = 0; x < pw; x += TS) {
        ctx.fillStyle = P.wl_castle
        ctx.fillRect(x, y, TS, TS)
        ctx.fillStyle = P.s2
        ctx.fillRect(x, y, TS, 1)
        ctx.fillRect(x, y + 7, TS, 1)
        ctx.fillRect(x, y, 1, TS)
        ctx.fillRect(x + 7, y, 1, TS)
      }
    }
    // Battlements
    ctx.fillStyle = P.wl_castle
    for (let x = 0; x < pw; x += TS * 2) {
      ctx.fillRect(x, wallTop - 4, TS, 4)
    }
  } else if (type === 'townhall') {
    // Municipal building
    for (let y = wallTop; y < ph - TS; y += TS) {
      for (let x = 0; x < pw; x += TS) {
        ctx.fillStyle = P.wl_white
        ctx.fillRect(x, y, TS, TS)
        ctx.fillStyle = P.s1
        ctx.fillRect(x, y, TS, 1)
        ctx.fillRect(x, y + 7, TS, 1)
      }
    }
    // Columns
    ctx.fillStyle = P.s3
    ctx.fillRect(TS - 2, wallTop, 3, ph - TS - wallTop)
    ctx.fillRect(pw - TS - 1, wallTop, 3, ph - TS - wallTop)
  } else {
    // Regular buildings
    const wallColor = type === 'bank' ? P.wl_brick : 
                      type === 'library' ? P.wl_wood :
                      type === 'shop' ? P.wl_wood :
                      type === 'home' ? P.wl_wood :
                      type === 'cottage' ? P.wl_wood :
                      type === 'pavilion' ? P.wl_stone :
                      type === 'police' ? P.wl_white :
                      type === 'hospital' ? P.wl_white :
                      type === 'watermill' ? P.wd1 :
                      type === 'farm' ? P.wd1 :
                      P.wl_wood
    for (let y = wallTop; y < ph - TS; y += TS) {
      for (let x = 0; x < pw; x += TS) {
        ctx.fillStyle = wallColor
        ctx.fillRect(x, y, TS, TS)
        ctx.fillStyle = shadeColor(wallColor, -15)
        ctx.fillRect(x, y, TS, 1)
        if (type === 'home' || type === 'cottage') {
          // Wood plank detail
          ctx.fillStyle = shadeColor(wallColor, 10)
          ctx.fillRect(x + 2, y + 2, TS - 4, TS - 4)
        }
      }
    }
  }
  
  // Windows
  const winRow = Math.floor(h / 2) * TS
  for (let x = TS; x < pw - TS; x += TS + 2) {
    // Window frame
    ctx.fillStyle = P.winFrame
    ctx.fillRect(x - 1, winRow - 1, TS + 2, TS + 2)
    // Glass
    ctx.fillStyle = P.win
    ctx.fillRect(x, winRow, TS, TS)
    // Window panes
    ctx.fillStyle = P.winFrame
    ctx.fillRect(x + TS/2 - 1, winRow, 2, TS)
    ctx.fillRect(x, winRow + TS/2 - 1, TS, 2)
    // Reflection
    ctx.fillStyle = 'rgba(255,255,255,0.3)'
    ctx.fillRect(x + 1, winRow + 1, 3, 3)
  }
  
  // Door
  const doorX = Math.floor(pw / 2) - 4
  const doorY = ph - TS - 8
  ctx.fillStyle = P.door
  ctx.fillRect(doorX, doorY, 8, 14)
  ctx.fillStyle = P.doorHandle
  ctx.fillRect(doorX + 5, doorY + 6, 2, 2)
  // Door frame
  ctx.fillStyle = P.wd2
  ctx.fillRect(doorX - 1, doorY - 1, 10, 16)
  ctx.fillRect(doorX, doorY - 2, 8, 2)
  
  // Roof
  drawRoof(ctx, type, pw, wallTop, h)
}

function drawRoof(ctx, type, pw, wallTop, h) {
  let r1, r2, r3
  switch (type) {
    case 'castle': r1 = P.r_blue1; r2 = P.r_blue2; r3 = P.r_blue3; break
    case 'townhall': r1 = P.r_blue1; r2 = P.r_blue2; r3 = P.r_blue3; break
    case 'bank': r1 = P.r_blue1; r2 = P.r_blue2; r3 = P.r_blue3; break
    case 'library': r1 = P.r_brown1; r2 = P.r_brown2; r3 = P.r_brown3; break
    case 'hospital': r1 = P.r_red1; r2 = P.r_red2; r3 = P.r_red3; break
    case 'police': r1 = P.r_blue1; r2 = P.r_blue2; r3 = P.r_blue3; break
    case 'pavilion': r1 = P.r_blue1; r2 = P.r_blue2; r3 = P.r_blue3; break
    case 'home': r1 = P.r_red1; r2 = P.r_red2; r3 = P.r_red3; break
    case 'cottage': r1 = P.r_green1; r2 = P.r_green2; r3 = P.r_green3; break
    case 'shop': r1 = P.r_blue1; r2 = P.r_blue2; r3 = P.r_blue3; break
    case 'watermill': r1 = P.r_brown1; r2 = P.r_brown2; r3 = P.r_brown3; break
    case 'farm': r1 = P.r_red1; r2 = P.r_red2; r3 = P.r_red3; break
    default: r1 = P.r_red1; r2 = P.r_red2; r3 = P.r_red3
  }
  
  // Triangular roof
  const roofH = h > 3 ? TS * 2 : TS + 4
  ctx.fillStyle = r1
  ctx.beginPath()
  ctx.moveTo(-2, wallTop)
  ctx.lineTo(pw / 2, wallTop - roofH)
  ctx.lineTo(pw + 2, wallTop)
  ctx.fill()
  
  // Roof shadow (right half)
  ctx.fillStyle = r2
  ctx.beginPath()
  ctx.moveTo(pw / 2, wallTop - roofH)
  ctx.lineTo(pw + 2, wallTop)
  ctx.lineTo(pw / 2, wallTop)
  ctx.fill()
  
  // Roof highlights
  ctx.fillStyle = r3
  ctx.fillRect(pw / 2 - 1, wallTop - roofH, 2, roofH - 2)
  
  // Chimney for homes
  if (type === 'home' || type === 'cottage') {
    ctx.fillStyle = P.s2
    ctx.fillRect(pw - TS - 4, wallTop - roofH + 4, 8, roofH - 4)
    ctx.fillStyle = P.s1
    ctx.fillRect(pw - TS - 5, wallTop - roofH + 2, 10, 4)
  }
  
  // Castle towers
  if (type === 'castle') {
    // Corner towers
    for (let tx of [0, pw - TS]) {
      ctx.fillStyle = P.wl_castle
      ctx.fillRect(tx, wallTop - TS * 3, TS, TS * 3)
      ctx.fillStyle = P.s2
      ctx.fillRect(tx, wallTop - TS * 3, TS, 1)
      ctx.fillRect(tx, wallTop - TS * 2, TS, 1)
      // Tower roof (cone)
      ctx.fillStyle = r1
      ctx.beginPath()
      ctx.moveTo(tx - 2, wallTop - TS * 3)
      ctx.lineTo(tx + TS / 2, wallTop - TS * 4.5)
      ctx.lineTo(tx + TS + 2, wallTop - TS * 3)
      ctx.fill()
      // Tower windows
      ctx.fillStyle = P.win
      ctx.fillRect(tx + 4, wallTop - TS * 2, 8, 8)
      ctx.fillStyle = P.winFrame
      ctx.fillRect(tx + 7, wallTop - TS * 2, 2, 8)
    }
    // Center gatehouse
    const gw = TS * 2
    ctx.fillStyle = P.wl_castle
    ctx.fillRect(pw/2 - gw/2, wallTop - TS * 2, gw, TS * 2)
    // Gate
    ctx.fillStyle = P.door
    ctx.fillRect(pw/2 - 4, wallTop - TS, 8, TS)
    ctx.fillStyle = P.doorHandle
    ctx.fillRect(pw/2 + 1, wallTop - TS/2, 2, 2)
  }
}

// ─── CHARACTER PRESETS ───
export const CHARACTERS = {
  agent: {
    name: 'Agent', skinColor: P.sk1, hairColor: P.hr_black,
    clothColor: P.cl_blue, hairStyle: 'short',
  },
  agent_f: {
    name: 'Agent', skinColor: P.sk1, hairColor: P.hr_blonde,
    clothColor: P.cl_red, hairStyle: 'long',
  },
  agent_robot: {
    name: 'Agent', skinColor: '#8899AA', hairColor: '#AABBCC',
    clothColor: '#667788', hairStyle: 'robot',
  },
  guard: {
    name: 'Guard', skinColor: P.sk2, hairColor: P.hr_black,
    clothColor: P.cl_grey, hairStyle: 'cap',
  },
  mayor: {
    name: 'Mayor Pixel', skinColor: P.sk1, hairColor: P.hr_red,
    clothColor: P.cl_purple, hairStyle: 'ponytail',
  },
  banker: {
    name: 'Banker', skinColor: P.sk2, hairColor: P.hr_silver,
    clothColor: P.cl_black, hairStyle: 'slick',
  },
  teacher: {
    name: 'Professor', skinColor: P.sk3, hairColor: P.hr_brown,
    clothColor: P.cl_green, hairStyle: 'long',
  },
  shopkeeper: {
    name: 'Shopkeeper', skinColor: P.sk1, hairColor: P.hr_blonde,
    clothColor: P.cl_orange, hairStyle: 'short',
  },
  farmer: {
    name: 'Farmer', skinColor: P.sk4, hairColor: P.hr_brown,
    clothColor: P.cl_green, hairStyle: 'hat',
  },
  // NPC developers for the business building
  dev_lead: {
    name: 'Alex Dev', skinColor: P.sk3, hairColor: P.hr_black,
    clothColor: P.cl_blue, hairStyle: 'short',
  },
  dev_junior: {
    name: 'Sam Code', skinColor: P.sk2, hairColor: P.hr_brown,
    clothColor: P.cl_green, hairStyle: 'cap',
  },
  dev_senior: {
    name: 'Jordan Bit', skinColor: P.sk1, hairColor: P.hr_blonde,
    clothColor: P.cl_purple, hairStyle: 'long',
  },
  designer: {
    name: 'Riley Pixel', skinColor: P.sk4, hairColor: P.hr_red,
    clothColor: P.cl_red, hairStyle: 'ponytail',
  },
  pm: {
    name: 'Casey Plan', skinColor: P.sk2, hairColor: P.hr_silver,
    clothColor: P.cl_black, hairStyle: 'slick',
  },
  devops: {
    name: 'Morgan Deploy', skinColor: P.sk1, hairColor: P.hr_black,
    clothColor: P.cl_grey, hairStyle: 'robot',
  },
  intern: {
    name: 'Taylor Learn', skinColor: P.sk3, hairColor: P.hr_blonde,
    clothColor: P.cl_orange, hairStyle: 'short',
  },
}

// ─── BUILDING DEFINITIONS ───
export const BUILDINGS = {
  home:     { w: 3, h: 3, type: 'home',     label: '' },
  cottage:  { w: 2, h: 2, type: 'cottage',  label: '' },
  shop:     { w: 3, h: 2, type: 'shop',     label: '' },
  bank:     { w: 4, h: 3, type: 'bank',     label: '' },
  library:  { w: 4, h: 3, type: 'library',  label: '' },
  townhall: { w: 5, h: 4, type: 'townhall', label: '' },
  pavilion: { w: 5, h: 3, type: 'pavilion', label: '' },
  castle:   { w: 6, h: 5, type: 'castle',   label: '' },
  watermill:{ w: 2, h: 3, type: 'watermill',label: '' },
  farm:     { w: 4, h: 2, type: 'farm',     label: '' },
  police:   { w: 3, h: 2, type: 'police',   label: '' },
  hospital: { w: 3, h: 2, type: 'hospital', label: '' },
  business: { w: 5, h: 4, type: 'business', label: 'Bitsburgh Dev Co.' },
}

// ─── BITSBURG MAP ───
const MW = 64, MH = 48

function buildMap() {
  const m = []
  for (let y = 0; y < MH; y++) {
    m[y] = []
    for (let x = 0; x < MW; x++) {
      m[y][x] = T.GRASS
    }
  }
  
  // Main roads
  for (let x = 0; x < MW; x++) { m[23][x] = T.PATH; m[24][x] = T.PATH }
  for (let y = 0; y < MH; y++) { m[y][31] = T.PATH; m[y][32] = T.PATH }
  for (let x = 0; x < MW; x++) m[11][x] = T.PATH
  for (let x = 0; x < MW; x++) m[36][x] = T.PATH
  for (let y = 0; y < MH; y++) { m[y][15] = T.PATH; m[y][48] = T.PATH }
  
  // Water
  for (let x = 10; x < 54; x++) for (let y = 42; y < 46; y++) m[y][x] = T.WATER
  for (let y = 8; y < 11; y++) for (let x = 10; x < 14; x++) m[y][x] = T.WATER_DEEP
  for (let y = 30; y < 33; y++) for (let x = 42; x < 46; x++) m[y][x] = T.WATER_DEEP
  
  // Bridges
  for (let x = 30; x < 34; x++) for (let y = 42; y < 46; y++) m[y][x] = T.BRIDGE
  for (let y = 8; y < 11; y++) { m[y][14] = T.BRIDGE; m[y][15] = T.BRIDGE }
  
  // Trees (scattered)
  const treePositions = [
    [1,3],[3,1],[8,3],[12,1],[1,8],[13,8],[2,10],[14,10],
    [37,2],[40,3],[45,2],[50,3],[55,5],[37,8],[42,8],[48,10],[55,10],
    [38,15],[45,15],[50,18],[55,15],[37,20],[42,20],[50,22],[58,18],
    [1,26],[3,28],[8,26],[12,28],[1,35],[5,35],[10,35],
    [37,36],[40,38],[48,36],[55,38],[58,35],
  ]
  treePositions.forEach(([x,y]) => { if (y < MH && x < MW) m[y][x] = T.TREE })
  
  // Flowers
  const flowerPositions = [
    [16,12],[17,13],[20,12],[25,13],[18,14],[22,12],
    [38,12],[40,13],[44,12],[46,13],[42,14],
    [16,37],[18,38],[22,37],[25,38],
    [38,25],[40,26],[44,25],[46,26],
  ]
  flowerPositions.forEach(([x,y]) => { if (y < MH && x < MW) m[y][x] = T.FLOWER })
  
  // Fountain in center plaza
  m[22][30] = T.FOUNTAIN; m[22][33] = T.FOUNTAIN
  m[25][30] = T.FOUNTAIN; m[25][33] = T.FOUNTAIN
  
  // Crops in farm area
  for (let x = 3; x < 7; x++) for (let y = 39; y < 41; y++) m[y][x] = T.CROP
  
  // Fences around farm
  for (let x = 2; x < 8; x++) { m[38][x] = T.FENCE; m[41][x] = T.FENCE }
  for (let y = 38; y < 42; y++) { m[y][2] = T.FENCE; m[y][7] = T.FENCE }
  
  // Fences in recreation
  for (let x = 37; x < 42; x++) { m[12][x] = T.FENCE; m[17][x] = T.FENCE }
  
  return m
}

const baseMap = buildMap()

// Building placements
const buildingPlacements = [
  // Residential (NW)
  { x: 2, y: 2, type: 'home' }, { x: 6, y: 2, type: 'home' },
  { x: 10, y: 2, type: 'cottage' }, { x: 2, y: 6, type: 'home' },
  { x: 6, y: 6, type: 'cottage' }, { x: 10, y: 6, type: 'home' },
  // Municipal (N-center)
  { x: 16, y: 2, type: 'police' }, { x: 20, y: 1, type: 'townhall' },
  { x: 26, y: 2, type: 'hospital' },
  // Business (SW)
  { x: 2, y: 26, type: 'shop' }, { x: 6, y: 26, type: 'shop' },
  { x: 10, y: 26, type: 'shop' }, { x: 2, y: 30, type: 'shop' },
  { x: 6, y: 30, type: 'bank' }, { x: 10, y: 30, type: 'shop' },
  // Banking & Education (SE)
  { x: 36, y: 26, type: 'bank' }, { x: 42, y: 26, type: 'library' },
  { x: 48, y: 26, type: 'shop' }, { x: 36, y: 30, type: 'shop' },
  { x: 42, y: 30, type: 'pavilion' },
  // Castle
  { x: 24, y: 13, type: 'castle' },
  // Farm
  { x: 2, y: 38, type: 'farm' },
  // Watermill
  { x: 18, y: 6, type: 'watermill' },
  // Business building (Dev company) - placed in business district
  { x: 18, y: 26, type: 'business' },
]

// Fill building tiles
buildingPlacements.forEach(b => {
  const def = BUILDINGS[b.type]
  if (!def) return
  for (let dy = 0; dy < def.h; dy++) {
    for (let dx = 0; dx < def.w; dx++) {
      const ty = b.y + dy, tx = b.x + dx
      if (ty >= 0 && ty < MH && tx >= 0 && tx < MW) {
        baseMap[ty][tx] = T.WALL
      }
    }
  }
})

const buildings = buildingPlacements.map(b => {
  const def = BUILDINGS[b.type]
  return { x: b.x, y: b.y, w: def.w, h: def.h, type: b.type, label: def.label,
    zone: getZone(b.x, b.y) }
})

function getZone(x, y) {
  if (x < 15 && y < 12) return 'residential'
  if (x >= 15 && x < 25 && y >= 24 && y < 32) return 'business_dev'
  if (x < 15 && y > 25) return 'business'
  if (x > 15 && x < 30 && y < 12) return 'municipal'
  if (x > 35 && y < 22) return 'recreation'
  if (x > 35 && y > 25) return 'banking'
  if (x > 20 && x < 33 && y > 12 && y < 22) return 'castle'
  return 'outskirts'
}

export const BITSBURG_MAP = baseMap
export const BITSBURG_BUILDINGS = buildings
export const MAP_WIDTH = MW
export const MAP_HEIGHT = MH
export const TILE_SIZE = TS

// ─── RENDER HELPERS ───
// Draw a tile at screen position
export function drawTile(ctx, type, x, y, s) {
  const tile = getTileCanvas(type)
  ctx.drawImage(tile, x, y, TS * s, TS * s)
}

// Draw a character at screen position
export function drawChar(ctx, config, x, y, dir, frame, s) {
  const sprite = getCharCanvas(config, dir, frame)
  ctx.drawImage(sprite, x, y, 16 * s, 24 * s)
}

// Draw a building at screen position
export function drawBuild(ctx, type, w, h, x, y, s) {
  const sprite = getBuildingCanvas(type, w, h)
  ctx.drawImage(sprite, x, y, w * TS * s, h * TS * s)
}
