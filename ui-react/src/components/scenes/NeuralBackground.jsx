import React, { useRef, useMemo, useEffect, useState } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import * as THREE from 'three';

// --- Configuration
const NODE_COUNT = 700;                // Total neurons in the field
const MAX_RADIUS = 7.0;               // Maximum expansion radius
const CONNECTION_DISTANCE = 2.2;      // Max distance to draw a connection
const COLORS = ['#06959d', '#853cbf', '#03c165'];
const EXPAND_SPEED = 0.4;             // Units per second (radius change) — slowed 50%
const HOLD_TIME = 1.8;                // Seconds to pause at full/empty — slowed 50%

// --- Helper: random position inside a sphere
const randomPointInSphere = (radius) => {
  const u = Math.random();
  const v = Math.random();
  const theta = 2 * Math.PI * u;
  const phi = Math.acos(2 * v - 1);
  const r = radius * Math.cbrt(Math.random()); // uniform distribution
  return new THREE.Vector3(
    r * Math.sin(phi) * Math.cos(theta),
    r * Math.sin(phi) * Math.sin(theta),
    r * Math.cos(phi)
  );
};

// --- Generate nodes (positions + colors)
const generateNodes = () => {
  const nodes = [];
  // Center node at origin (always present)
  nodes.push({
    id: 0,
    position: new THREE.Vector3(0, 0, 0),
    color: COLORS[0],
    radius: 0.32,
  });
  // Generate random nodes within MAX_RADIUS, avoiding origin
  for (let i = 1; i < NODE_COUNT; i++) {
    let pos;
    let tooClose = true;
    // Avoid overlapping exactly with origin or other nodes too much (optional)
    while (tooClose) {
      pos = randomPointInSphere(MAX_RADIUS - 0.5);
      tooClose = pos.length() < 0.6;
    }
    const randomColor = COLORS[Math.floor(Math.random() * COLORS.length)];
    nodes.push({
      id: i,
      position: pos,
      color: randomColor,
      radius: 0.22 + Math.random() * 0.1,
    });
  }
  return nodes;
};

// --- Precompute connections between nodes that are within distance
const computeConnections = (nodes, maxDist) => {
  const connections = [];
  for (let i = 0; i < nodes.length; i++) {
    for (let j = i + 1; j < nodes.length; j++) {
      const dist = nodes[i].position.distanceTo(nodes[j].position);
      if (dist < maxDist) {
        connections.push([i, j]);
      }
    }
  }
  return connections;
};

// --- Component for a single glowing neuron
const Neuron = ({ position, color, radius, active, waveProgress }) => {
  const meshRef = useRef();
  const materialRef = useRef();
  
  useFrame(({ clock }) => {
    if (meshRef.current) {
      // Pulsation based on activation and time
      const pulse = active ? 0.9 + Math.sin(clock.getElapsedTime() * 8) * 0.15 : 0.5;
      // Scale also influenced by wave proximity (firing effect)
      const scale = (active ? 0.8 + waveProgress * 0.5 : 0.3) * pulse;
      meshRef.current.scale.setScalar(scale);
      // Emissive intensity
      if (materialRef.current) {
        const intensity = active ? 0.6 + Math.sin(clock.getElapsedTime() * 10) * 0.3 : 0.1;
        materialRef.current.emissiveIntensity = intensity;
      }
    }
  });
  
  return (
    <mesh ref={meshRef} position={position}>
      <sphereGeometry args={[radius, 32, 32]} />
      <meshStandardMaterial
        ref={materialRef}
        color={color}
        emissive={color}
        emissiveIntensity={0.4}
        roughness={0.3}
        metalness={0.6}
        transparent
        opacity={active ? 0.56 : 0.15}
      />
    </mesh>
  );
};

// --- Dynamic connections: only draw lines where both endpoints are active
const DynamicConnections = ({ nodes, connections, activeMap }) => {
  const lineRef = useRef();
  const geometryRef = useRef();
  const positionsArray = useMemo(() => new Float32Array(connections.length * 2 * 3), [connections.length]);
  const colorArray = useMemo(() => new Float32Array(connections.length * 2 * 3), [connections.length]);

  useEffect(() => {
    if (geometryRef.current) {
      geometryRef.current.setAttribute('position', new THREE.BufferAttribute(positionsArray, 3));
      geometryRef.current.setAttribute('color', new THREE.BufferAttribute(colorArray, 3));
    }
  }, [positionsArray, colorArray]);

  useFrame(() => {
    if (!geometryRef.current) return;
    let idx = 0;
    let visibleCount = 0;
    for (let conn of connections) {
      const [idA, idB] = conn;
      const activeA = activeMap.current[idA];
      const activeB = activeMap.current[idB];
      if (activeA && activeB) {
        const posA = nodes[idA].position;
        const posB = nodes[idB].position;
        positionsArray[idx] = posA.x;
        positionsArray[idx+1] = posA.y;
        positionsArray[idx+2] = posA.z;
        positionsArray[idx+3] = posB.x;
        positionsArray[idx+4] = posB.y;
        positionsArray[idx+5] = posB.z;
        // Color interpolation between the two nodes' colors
        const colorA = new THREE.Color(nodes[idA].color);
        const colorB = new THREE.Color(nodes[idB].color);
        const avgColor = colorA.lerp(colorB, 0.5);
        // Brightness varies with a sine wave for "firing" effect
        const intensity = 0.5 + Math.sin(Date.now() * 0.008) * 0.3;
        avgColor.multiplyScalar(intensity);
        for (let i = 0; i < 2; i++) {
          colorArray[idx + i*3] = avgColor.r;
          colorArray[idx + i*3 + 1] = avgColor.g;
          colorArray[idx + i*3 + 2] = avgColor.b;
        }
        idx += 6;
        visibleCount++;
      } else {
        // Make line invisible by setting positions to zero (or skip, but we need to keep stride)
        // Simpler: set positions to a point far away
        positionsArray[idx] = 0; positionsArray[idx+1] = 0; positionsArray[idx+2] = 0;
        positionsArray[idx+3] = 0; positionsArray[idx+4] = 0; positionsArray[idx+5] = 0;
        idx += 6;
      }
    }
    geometryRef.current.attributes.position.needsUpdate = true;
    geometryRef.current.attributes.color.needsUpdate = true;
    // Adjust draw range to only visible lines (performance)
    if (lineRef.current) {
      lineRef.current.geometry.setDrawRange(0, visibleCount * 2);
    }
  });

  return (
    <lineSegments ref={lineRef}>
      <bufferGeometry ref={geometryRef} attach="geometry" />
      <lineBasicMaterial vertexColors={true} linewidth={1} transparent opacity={0.4} />
    </lineSegments>
  );
};

// --- Particle system for firing synapses during expansion wave
const FiringParticles = ({ nodes, waveRadius, maxRadius }) => {
  const particleCount = 1500;
  const particlesRef = useRef();
  const positionsRef = useRef(new Float32Array(particleCount * 3));
  const velocitiesRef = useRef([]);
  const activeTimeRef = useRef(0);

  useEffect(() => {
    const positions = new Float32Array(particleCount * 3);
    const velocities = [];
    for (let i = 0; i < particleCount; i++) {
      // Random initial positions within sphere
      const radius = Math.random() * maxRadius;
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      positions[i*3] = Math.sin(phi) * Math.cos(theta) * radius;
      positions[i*3+1] = Math.sin(phi) * Math.sin(theta) * radius;
      positions[i*3+2] = Math.cos(phi) * radius;
      velocities.push({
        x: (Math.random() - 0.5) * 0.025,
        y: (Math.random() - 0.5) * 0.025,
        z: (Math.random() - 0.5) * 0.025,
      });
    }
    positionsRef.current = positions;
    velocitiesRef.current = velocities;
    if (particlesRef.current) {
      particlesRef.current.geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    }
  }, [maxRadius]);

  useFrame((_, delta) => {
    if (!particlesRef.current) return;
    const positions = particlesRef.current.geometry.attributes.position.array;
    // Intensity peaks when wave is expanding and around mid-radius
    const intensity = Math.sin(Math.PI * waveRadius / maxRadius) * (waveRadius < maxRadius ? 1 : 0);
    
    for (let i = 0; i < particleCount; i++) {
      // Outward push when wave is active
      if (intensity > 0.1) {
        const dir = new THREE.Vector3(positions[i*3], positions[i*3+1], positions[i*3+2]).normalize();
        const speed = 0.045 * intensity;
        positions[i*3] += dir.x * speed + velocitiesRef.current[i].x * delta;
        positions[i*3+1] += dir.y * speed + velocitiesRef.current[i].y * delta;
        positions[i*3+2] += dir.z * speed + velocitiesRef.current[i].z * delta;
        
        // Reset particles that go too far or drift inside
        const mag = Math.hypot(positions[i*3], positions[i*3+1], positions[i*3+2]);
        if (mag > maxRadius + 1.5 || mag < 0.5) {
          const radius = Math.random() * maxRadius * 0.8;
          const theta = Math.random() * Math.PI * 2;
          const phi = Math.acos(2 * Math.random() - 1);
          positions[i*3] = Math.sin(phi) * Math.cos(theta) * radius;
          positions[i*3+1] = Math.sin(phi) * Math.sin(theta) * radius;
          positions[i*3+2] = Math.cos(phi) * radius;
        }
      } else {
        // Drift slowly back to center during collapse
        positions[i*3] *= 0.99;
        positions[i*3+1] *= 0.99;
        positions[i*3+2] *= 0.99;
      }
    }
    particlesRef.current.geometry.attributes.position.needsUpdate = true;
    particlesRef.current.material.opacity = intensity * 0.32;
  });

  return (
    <points ref={particlesRef}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" args={[positionsRef.current, 3]} />
      </bufferGeometry>
      <pointsMaterial color="#03c165" size={0.07} transparent blending={THREE.AdditiveBlending} />
    </points>
  );
};

// --- Main scene with wave expansion / collapse
const NeuralScene = () => {
  const nodes = useMemo(() => generateNodes(), []);
  const connections = useMemo(() => computeConnections(nodes, CONNECTION_DISTANCE), [nodes]);
  const activeMap = useRef(new Array(nodes.length).fill(false));
  const waveRadiusRef = useRef(0);
  const timeStateRef = useRef({ phase: 'expanding', timer: 0 });

  // Update wave radius and active nodes based on state machine
  useFrame((_, delta) => {
    const state = timeStateRef.current;
    let radius = waveRadiusRef.current;
    let changed = false;

    switch (state.phase) {
      case 'expanding':
        radius += EXPAND_SPEED * delta;
        if (radius >= MAX_RADIUS) {
          radius = MAX_RADIUS;
          state.phase = 'hold_full';
          state.timer = 0;
        }
        break;
      case 'hold_full':
        state.timer += delta;
        if (state.timer >= HOLD_TIME) {
          state.phase = 'collapsing';
        }
        break;
      case 'collapsing':
        radius -= EXPAND_SPEED * delta;
        if (radius <= 0) {
          radius = 0;
          state.phase = 'hold_empty';
          state.timer = 0;
        }
        break;
      case 'hold_empty':
        state.timer += delta;
        if (state.timer >= HOLD_TIME) {
          state.phase = 'expanding';
        }
        break;
    }
    waveRadiusRef.current = radius;

    // Update active nodes: active if distance from origin <= radius
    for (let i = 0; i < nodes.length; i++) {
      const dist = nodes[i].position.length();
      const shouldBeActive = dist <= radius;
      if (activeMap.current[i] !== shouldBeActive) {
        activeMap.current[i] = shouldBeActive;
        changed = true;
      }
    }
  });

  // Compute wave progress (0..1) for firing effects
  const waveProgress = waveRadiusRef.current / MAX_RADIUS;

  return (
    <>
      <ambientLight intensity={0.35} />
      <pointLight position={[4, 5, 3]} intensity={0.7} color="#853cbf" />
      <pointLight position={[-3, 2, 5]} intensity={0.6} color="#06959d" />
      <pointLight position={[2, -3, 4]} intensity={0.5} color="#03c165" />

      {/* Render all neurons */}
      {nodes.map((node) => (
        <Neuron
          key={node.id}
          position={node.position}
          color={node.color}
          radius={node.radius}
          active={activeMap.current[node.id]}
          waveProgress={waveProgress}
        />
      ))}

      {/* Dynamic connections based on active nodes */}
      <DynamicConnections nodes={nodes} connections={connections} activeMap={activeMap} />

      {/* Particle system for firing effect */}
      <FiringParticles nodes={nodes} waveRadius={waveRadiusRef.current} maxRadius={MAX_RADIUS} />
    </>
  );
};

// --- Main App Component
const NeuralBackground = () => {
  return (
    <div style={{ width: '100vw', height: '100vh', position: 'fixed', top: 0, left: 0, zIndex: 0, pointerEvents: 'none', background: '#050510' }}>
      <Canvas
        camera={{ position: [0, 2, 10], fov: 55 }}
        style={{ background: 'radial-gradient(circle at center, #0a0a2a 0%, #020210 100%)', pointerEvents: 'none' }}
        eventSource={undefined}
        eventPrefix="client"
      >
        <OrbitControls
          enableZoom={false}
          enablePan={false}
          autoRotate
          autoRotateSpeed={0.6}
          enableDamping
          dampingFactor={0.05}
        />
        <NeuralScene />
      </Canvas>
    </div>
  );
};

export default NeuralBackground;
