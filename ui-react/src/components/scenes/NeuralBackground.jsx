import React, { useRef, useMemo, useEffect, useState } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import * as THREE from 'three';

// --- Simple auto-rotation (replaces OrbitControls to avoid drei dependency issues)
const AutoRotate = () => {
  const { camera } = useThree();
  const angleRef = useRef(0);
  
  useFrame((_, delta) => {
    angleRef.current += delta * 0.6; // autoRotateSpeed equivalent
    const radius = Math.sqrt(
      camera.position.x * camera.position.x + 
      camera.position.z * camera.position.z
    );
    const height = camera.position.y;
    camera.position.x = Math.cos(angleRef.current) * radius;
    camera.position.z = Math.sin(angleRef.current) * radius;
    camera.position.y = height;
    camera.lookAt(0, 0, 0);
  });
  
  return null;
};

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

// --- Instanced neuron renderer for performance (700 nodes)
const InstancedNeurons = ({ nodes, activeMap, waveProgress }) => {
  const meshRef = useRef();
  const count = nodes.length;

  useFrame(({ clock }) => {
    if (!meshRef.current) return;
    const t = clock.getElapsedTime();
    const dummy = new THREE.Object3D();
    const color = new THREE.Color();

    for (let i = 0; i < count; i++) {
      const node = nodes[i];
      const active = activeMap.current[i];
      const pulse = active ? 0.9 + Math.sin(t * 8 + i) * 0.15 : 0.5;
      const scale = (active ? 0.8 + waveProgress * 0.5 : 0.3) * pulse;

      dummy.position.copy(node.position);
      dummy.scale.setScalar(scale * node.radius);
      dummy.updateMatrix();
      meshRef.current.setMatrixAt(i, dummy.matrix);

      color.set(node.color);
      meshRef.current.setColorAt(i, color);
    }
    meshRef.current.instanceMatrix.needsUpdate = true;
    if (meshRef.current.instanceColor) meshRef.current.instanceColor.needsUpdate = true;
  });

  return (
    <instancedMesh ref={meshRef} args={[null, null, count]}>
      <sphereGeometry args={[1, 16, 16]} />
      <meshStandardMaterial
        emissiveIntensity={0.4}
        roughness={0.3}
        metalness={0.6}
        transparent
        opacity={0.8}
      />
    </instancedMesh>
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
        const colorA = new THREE.Color(nodes[idA].color);
        const colorB = new THREE.Color(nodes[idB].color);
        const avgColor = colorA.lerp(colorB, 0.5);
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
        positionsArray[idx] = 0; positionsArray[idx+1] = 0; positionsArray[idx+2] = 0;
        positionsArray[idx+3] = 0; positionsArray[idx+4] = 0; positionsArray[idx+5] = 0;
        idx += 6;
      }
    }
    geometryRef.current.attributes.position.needsUpdate = true;
    geometryRef.current.attributes.color.needsUpdate = true;
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

  useEffect(() => {
    const positions = new Float32Array(particleCount * 3);
    const velocities = [];
    for (let i = 0; i < particleCount; i++) {
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
    const intensity = Math.sin(Math.PI * waveRadius / maxRadius) * (waveRadius < maxRadius ? 1 : 0);

    for (let i = 0; i < particleCount; i++) {
      if (intensity > 0.1) {
        const dir = new THREE.Vector3(positions[i*3], positions[i*3+1], positions[i*3+2]).normalize();
        const speed = 0.045 * intensity;
        positions[i*3] += dir.x * speed + velocitiesRef.current[i].x * delta;
        positions[i*3+1] += dir.y * speed + velocitiesRef.current[i].y * delta;
        positions[i*3+2] += dir.z * speed + velocitiesRef.current[i].z * delta;

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

      {/* Instanced neurons for performance */}
      <InstancedNeurons nodes={nodes} activeMap={activeMap} waveProgress={waveProgress} />

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
        gl={{ antialias: true, alpha: true, preserveDrawingBuffer: false, powerPreference: 'high-performance' }}
      >
        <AutoRotate />
        <NeuralScene />
      </Canvas>
    </div>
  );
};

export default NeuralBackground;
