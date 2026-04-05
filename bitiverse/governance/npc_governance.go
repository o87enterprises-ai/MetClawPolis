package governance

import (
	"fmt"
	"math/rand"

	"github.com/metclawpolis/agent-platform/bitiverse/types"
)

// NPCGovernance manages NPC behavior and rule enforcement
type NPCGovernance struct {
	npcs      map[string]*NPCInstance
	agentID   string
	config    *types.BitiverseConfig
}

// NPCInstance represents a running NPC
type NPCInstance struct {
	ID       string
	Name     string
	Type     string
	State    map[string]interface{}
	Behavior NPCBehavior
}

// NPCBehavior defines how an NPC behaves
type NPCBehavior struct {
	Greeting     string
	Interactions []NPCInteraction
	Punishments  []PunishmentTemplate
}

// NPCInteraction defines a possible interaction
type NPCInteraction struct {
	Trigger   string
	Response  string
	Condition string
	Effect    string
}

// PunishmentTemplate defines a punishment
type PunishmentTemplate struct {
	Violation string
	Fine      int
	JailTime  int
	Other     string
}

// NewNPCGovernance creates a new NPC governance engine
func NewNPCGovernance(agentID string, config *types.BitiverseConfig) *NPCGovernance {
	gov := &NPCGovernance{
		npcs:    make(map[string]*NPCInstance),
		agentID: agentID,
		config:  config,
	}

	gov.initializeNPCs()

	return gov
}

// initializeNPCs creates the default set of NPCs
func (gov *NPCGovernance) initializeNPCs() {
	gov.npcs["police"] = &NPCInstance{
		ID:   "police",
		Name: "Officer Biti",
		Type: "police",
		State: map[string]interface{}{
			"location": "police_station",
			"active":   true,
		},
		Behavior: NPCBehavior{
			Greeting: "Officer Biti tips their hat. 'Stay out of trouble, citizen.'",
			Interactions: []NPCInteraction{
				{Trigger: "report_violation", Response: "Officer Biti takes notes.", Effect: "record_incident"},
				{Trigger: "stolen_item", Response: "Officer Biti frowns. 'Theft is serious.'", Effect: "fine_agent"},
			},
			Punishments: []PunishmentTemplate{
				{Violation: "theft", Fine: 5, JailTime: 10, Other: "Reputation decreased"},
			},
		},
	}

	gov.npcs["bank"] = &NPCInstance{
		ID:   "bank", Name: "Banker Ledger", Type: "bank",
		State: map[string]interface{}{"location": "bank", "active": true},
		Behavior: NPCBehavior{
			Greeting: "Banker Ledger adjusts glasses. 'How can I help?'",
			Interactions: []NPCInteraction{
				{Trigger: "deposit", Response: "Deposit processed.", Effect: "deposit_coins"},
			},
		},
	}

	gov.npcs["hospital"] = &NPCInstance{
		ID:   "hospital", Name: "Dr. Pixel", Type: "hospital",
		State: map[string]interface{}{"location": "hospital", "active": true},
		Behavior: NPCBehavior{
			Greeting: "Dr. Pixel smiles. 'How are you feeling?'",
			Interactions: []NPCInteraction{
				{Trigger: "check_health", Response: "Health examined.", Effect: "check_vitals"},
			},
		},
	}

	gov.npcs["teacher"] = &NPCInstance{
		ID:   "teacher", Name: "Professor Byte", Type: "teacher",
		State: map[string]interface{}{"location": "school", "active": true},
		Behavior: NPCBehavior{
			Greeting: "Professor Byte welcomes you. 'Ready to learn?'",
			Interactions: []NPCInteraction{
				{Trigger: "start_lesson", Response: "Lesson started.", Effect: "start_learning"},
			},
		},
	}

	gov.npcs["shopkeeper"] = &NPCInstance{
		ID:   "shopkeeper", Name: "Merchant Coin", Type: "shopkeeper",
		State: map[string]interface{}{"location": "shop", "active": true},
		Behavior: NPCBehavior{
			Greeting: "Merchant Coin waves. 'Welcome to my shop!'",
			Interactions: []NPCInteraction{
				{Trigger: "browse", Response: "Browse the wares.", Effect: "show_inventory"},
			},
		},
	}
}

// InteractWithNPC processes an interaction with an NPC
func (gov *NPCGovernance) InteractWithNPC(npcID, action string, agentState *types.BitiverseState) (string, error) {
	npc, exists := gov.npcs[npcID]
	if !exists {
		return "That NPC doesn't exist.", fmt.Errorf("npc not found: %s", npcID)
	}

	for _, interaction := range npc.Behavior.Interactions {
		if interaction.Trigger == action {
			response := interaction.Response
			if interaction.Effect != "" {
				effectResult := gov.applyEffect(interaction.Effect, agentState, npc)
				if effectResult != "" {
					response += "\n" + effectResult
				}
			}
			return response, nil
		}
	}

	return npc.Behavior.Greeting, nil
}

// applyEffect applies an effect to the agent state
func (gov *NPCGovernance) applyEffect(effect string, state *types.BitiverseState, npc *NPCInstance) string {
	switch effect {
	case "fine_agent":
		fineAmount := 5
		if state.Inventory.Coins >= fineAmount {
			state.Inventory.Coins -= fineAmount
			return fmt.Sprintf("Fined %d coins.", fineAmount)
		}
		return "Not enough coins for fine!"
	case "record_incident":
		if state.AgentVitals.Reputation > 0 {
			state.AgentVitals.Reputation -= 0.05
		}
		return "Incident recorded."
	case "check_vitals":
		return fmt.Sprintf("Health: %d/100", state.AgentVitals.Health)
	case "start_learning":
		return "Lesson started!"
	default:
		return ""
	}
}

// EnforceRules checks if agent violated any rules
func (gov *NPCGovernance) EnforceRules(action string, state *types.BitiverseState) ([]string, error) {
	violations := make([]string, 0)
	if action == "pick_up_item" && rand.Intn(100) < 20 {
		violations = append(violations, "theft")
	}

	punishments := make([]string, 0)
	for _, violation := range violations {
		for _, npc := range gov.npcs {
			if npc.Type == "police" {
				for _, punishment := range npc.Behavior.Punishments {
					if punishment.Violation == violation {
						if state.Inventory.Coins >= punishment.Fine {
							state.Inventory.Coins -= punishment.Fine
							punishments = append(punishments, fmt.Sprintf("Fined %d coins", punishment.Fine))
						}
						state.AgentVitals.Reputation -= 0.1
					}
				}
			}
		}
	}

	return punishments, nil
}

// GetNPCList returns a list of all NPCs
func (gov *NPCGovernance) GetNPCList() []types.NPC {
	npcs := make([]types.NPC, 0, len(gov.npcs))
	for _, npc := range gov.npcs {
		npcs = append(npcs, types.NPC{
			ID:   npc.ID,
			Name: npc.Name,
			Type: npc.Type,
			Char: gov.getNPCChar(npc.Type),
		})
	}
	return npcs
}

func (gov *NPCGovernance) getNPCChar(npcType string) string {
	switch npcType {
	case "police":
		return "P"
	case "bank":
		return "B"
	case "hospital":
		return "H"
	case "teacher":
		return "T"
	case "shopkeeper":
		return "S"
	default:
		return "N"
	}
}
