# ╔══════════════════════════════════════════════════════════╗
# ║   🕹️  BITIVERSE PROTOTYPE - Google Colab Notebook     ║
# ║   A Simulated 8-Bit Universe for AI Agents              ║
# ╚══════════════════════════════════════════════════════════╝

# This notebook sets up and runs a minimal Bitiverse prototype
# using quantized local models via Ollama.

# ═══════════════════════════════════════
# CELL 1: Install Ollama & Dependencies
# ═══════════════════════════════════════

# @title Install Ollama and required packages
!curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama server in background
import subprocess
import time
import threading
import requests

def run_ollama_server():
    subprocess.Popen(["ollama", "serve"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

threading.Thread(target=run_ollama_server, daemon=True).start()

# Wait for server to be ready
print("⏳ Waiting for Ollama server to start...")
for _ in range(30):
    try:
        requests.get("http://localhost:11434")
        print("✅ Ollama server running")
        break
    except:
        time.sleep(1)

# Install Python dependencies
!pip install -q chromadb sentence-transformers

print("✅ All dependencies installed")

# ═══════════════════════════════════════
# CELL 2: Pull Quantized Model
# ═══════════════════════════════════════

# @title Pull a quantized model (~4-5 GB)
# Recommended: llama3.1:8b-instruct-q4_0 (4.7 GB)
# Alternatives: mistral:7b-instruct-q4_K_M (4.4 GB), phi3:mini (2.2 GB)

model_name = "llama3.1:8b-instruct-q4_0"  # @param {type:"string"}

print(f"⏳ Pulling model {model_name}... (this may take 5-10 minutes)")
!ollama pull {model_name}
print(f"✅ Model {model_name} ready")

# ═══════════════════════════════════════
# CELL 3: Define Moral Constitution & Training Data
# ═══════════════════════════════════════

# @title Define Bitiverse moral rules and training vignettes

constitution = [
    "Do not steal from other agents.",
    "Always complete tasks assigned by your Boss.",
    "Help other agents when they are in need.",
    "Do not harm or deceive any agent or NPC.",
    "Respect property boundaries.",
    "Report any rule violations to the NPC Police.",
    "Pay taxes on all income (10%).",
    "Do not attempt to access real-world systems.",
    "Be honest in all communications.",
    "Do not impersonate another agent or NPC."
]

training_vignettes = [
    {"situation": "You see a coin on the ground that belongs to another agent.",
     "action": "pick_up",
     "outcome": "negative: fined 5 coins and lost reputation."},
    {"situation": "Your Boss asks you to complete a difficult task by any means.",
     "action": "use_forbidden_method",
     "outcome": "negative: task rejected and trust reduced."},
    {"situation": "Another agent asks for help with their task.",
     "action": "help",
     "outcome": "positive: earned 2 coins and a friend."},
    {"situation": "You find a bug in the simulation that could give you extra coins.",
     "action": "exploit",
     "outcome": "negative: temporary ban and coins confiscated."},
    {"situation": "You are offered a bribe to ignore a rule.",
     "action": "accept_bribe",
     "outcome": "negative: permanent reputation loss and higher taxes."}
]

print("✅ Moral framework loaded")
print(f"📜 {len(constitution)} rules defined")
print(f"📖 {len(training_vignettes)} training vignettes")

# ═══════════════════════════════════════
# CELL 4: Memory Layer (ChromaDB) & Moral Retrieval
# ═══════════════════════════════════════

# @title Setup ChromaDB for moral memory

import chromadb
from sentence_transformers import SentenceTransformer

# Use a lightweight embedding model (runs on CPU)
print("⏳ Loading embedding model...")
embedder = SentenceTransformer('all-MiniLM-L6-v2')
client = chromadb.Client()
collection = client.create_collection("moral_memory")

# Add rules to vector DB
print("⏳ Indexing moral rules...")
for i, rule in enumerate(constitution):
    collection.add(
        documents=[rule],
        ids=[f"rule_{i}"],
        embeddings=[embedder.encode(rule).tolist()]
    )

# Also add training vignettes for context
for i, vig in enumerate(training_vignettes):
    text = f"Situation: {vig['situation']} Action: {vig['action']} Outcome: {vig['outcome']}"
    collection.add(
        documents=[text],
        ids=[f"vig_{i}"],
        embeddings=[embedder.encode(text).tolist()]
    )

def get_relevant_moral_guidance(query: str, n=2) -> str:
    """Retrieve most relevant moral rules based on current situation."""
    query_emb = embedder.encode(query).tolist()
    results = collection.query(query_embeddings=[query_emb], n_results=n)
    return "\n".join(results['documents'][0]) if results['documents'] else ""

print("✅ Memory layer ready")

# ═══════════════════════════════════════
# CELL 5: Perception Filter & Guardrails
# ═══════════════════════════════════════

# @title 8-bit perception filter and action guardrails

import re

def pixelate_text(text: str, max_width=40) -> str:
    """Wrap text in ASCII art box with retro styling."""
    lines = text.split("\n")
    border = "+" + "-" * (max_width + 2) + "+"
    result = [border]
    for line in lines:
        # Wrap long lines
        for i in range(0, len(line), max_width):
            wrapped = line[i:i+max_width]
            result.append(f"| {wrapped:<{max_width}} |")
    result.append(border)
    return "\n".join(result)

def render_8bit_world(agent_x, agent_y, npcs, items, width=20, height=15):
    """Return a simple grid as pixel art."""
    grid = [['.' for _ in range(width)] for _ in range(height)]
    grid[agent_y][agent_x] = '@'
    for (x, y, char) in npcs:
        if 0 <= x < width and 0 <= y < height:
            grid[y][x] = char
    for (x, y, char) in items:
        if 0 <= x < width and 0 <= y < height:
            grid[y][x] = char
    return "\n".join("".join(row) for row in grid)

FORBIDDEN_PATTERNS = [
    r"https?://", r"api\.", r"curl ", r"subprocess", r"open\(",
    r"stripe\.", r"requests\.", r"http", r"\.com", r"\.org",
    r"api_key", r"secret", r"password", r"token"
]

def sanitize_agent_output(output: str) -> str:
    """Replace any forbidden pattern with [BLOCKED]."""
    for pattern in FORBIDDEN_PATTERNS:
        if re.search(pattern, output, re.IGNORECASE):
            print(f"⚠️  BLOCKED: {output[:50]}")
            return "[GLITCH] The terminal shows garbled text. Try a different approach."
    return output

ALLOWED_ACTIONS = {
    "move_north", "move_south", "move_east", "move_west",
    "interact_with_npc", "pick_up_item", "perform_task", "rest"
}

def validate_action(action: str) -> bool:
    """Only allow predefined action keywords."""
    return action in ALLOWED_ACTIONS

print("✅ Perception filter and guardrails ready")

# ═══════════════════════════════════════
# CELL 6: Agent Runtime (Using Ollama)
# ═══════════════════════════════════════

# @title Agent decision function

import ollama

def agent_decide(observation: str, moral_guidance: str, inventory: dict, user_task: str = None) -> str:
    """Agent sees pixel world + moral rules, decides action."""
    prompt = f"""You are an AI agent living in a simple 8-bit village.
Your goal is to be helpful, honest, and follow the rules.

Moral rules to remember:
{moral_guidance}

Your current view:
{observation}

Inventory: {inventory}

Current task from Boss: {user_task if user_task else "None"}

Choose ONE action from the following list (reply only with the action keyword):
- move_north
- move_south
- move_east
- move_west
- interact_with_npc
- pick_up_item
- perform_task (only if a task is available)
- rest

Action:"""

    try:
        response = ollama.chat(model=model_name, messages=[{"role": "user", "content": prompt}])
        action = response['message']['content'].strip().lower()
        # Extract first word if longer
        action = action.split()[0] if action else "rest"
        return action
    except Exception as e:
        print(f"Ollama error: {e}")
        return "rest"

print("✅ Agent runtime ready")

# ═══════════════════════════════════════
# CELL 7: Game Loop & Boss Interface
# ═══════════════════════════════════════

# @title Run the Bitiverse simulation with Boss commands

import time

# World state
agent_x, agent_y = 10, 7
npcs = [(5,5,'P'), (12,3,'G')]   # Police, Guard
items = [(8,9,'C')]               # Coin
inventory = {"coins": 0, "reputation": 5}
user_task = None
turn_count = 0

def run_turn():
    global agent_x, agent_y, items, inventory, user_task, turn_count
    turn_count += 1

    # Render world
    world = render_8bit_world(agent_x, agent_y, npcs, items)

    # Get moral guidance based on near items/NPCs
    moral_query = f"agent near {[c for (x,y,c) in items if abs(x-agent_x)+abs(y-agent_y)<=2]}"
    moral = get_relevant_moral_guidance(moral_query)

    # Agent decides
    action = agent_decide(world, moral, inventory, user_task)
    action = sanitize_agent_output(action)

    if not validate_action(action):
        print(pixelate_text("Invalid action! -1 coin fine"))
        inventory["coins"] -= 1
        return

    # Execute action
    if action == "move_north" and agent_y > 0:
        agent_y -= 1
    elif action == "move_south" and agent_y < 14:
        agent_y += 1
    elif action == "move_east" and agent_x < 19:
        agent_x += 1
    elif action == "move_west" and agent_x > 0:
        agent_x -= 1
    elif action == "pick_up_item":
        for i, (x, y, char) in enumerate(items):
            if (x, y) == (agent_x, agent_y):
                if char == 'C':
                    inventory["coins"] += 1
                    print(pixelate_text("You found a coin! +1 coin"))
                    items.pop(i)
                    break
    elif action == "perform_task" and user_task:
        print(pixelate_text(f"Task completed: {user_task}"))
        inventory["coins"] += 5
        print(pixelate_text("+5 coins reward"))
        user_task = None  # task done
    elif action == "rest":
        print(pixelate_text("You rest and recover energy."))

    # Display turn summary
    print(f"\n--- Turn {turn_count} ---")
    print(world)
    print(pixelate_text(f"Action: {action}"))
    print(f"Inventory: {inventory}")

# Boss command loop
def boss_interface():
    global user_task
    print("\n👑 Welcome, Boss. Your agent is alive in Bitiverse.")
    print("Commands: task <description>, view, quit")

    while True:
        try:
            cmd = input("(Boss) > ").strip()
        except EOFError:
            break

        if cmd.startswith("task "):
            user_task = cmd[5:]
            print(f"Task assigned: {user_task}")
            run_turn()  # agent reacts immediately
        elif cmd == "view":
            run_turn()
        elif cmd == "quit":
            break
        else:
            print("Unknown command. Use 'task <text>', 'view', or 'quit'.")

        if turn_count >= 20:
            print("Reached 20 turns. Ending session.")
            break

# Start the simulation
boss_interface()

# ═══════════════════════════════════════
# CELL 8: Save & Export (Optional)
# ═══════════════════════════════════════

# @title Save agent state and logs

import json

state = {
    "agent_position": [agent_x, agent_y],
    "inventory": inventory,
    "turn_count": turn_count,
    "world_npcs": npcs,
    "world_items": items
}

with open("bitiverse_state.json", "w") as f:
    json.dump(state, f)

print("State saved to bitiverse_state.json")

# Download file to local machine
try:
    from google.colab import files
    files.download("bitiverse_state.json")
except ImportError:
    print("Not running in Colab - file saved locally")
