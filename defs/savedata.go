package defs

type SystemSaveData struct {
	TrainerId          int                `json:"trainerId"`
	SecretId           int                `json:"secretId"`
	Gender             int                `json:"gender"`
	DexData            DexData            `json:"dexData"`
	StarterMoveData    StarterMoveData    `json:"starterMoveData"`
	StarterEggMoveData StarterEggMoveData `json:"starterEggMoveData"`
	GameStats          GameStats          `json:"gameStats"`
	Unlocks            Unlocks            `json:"unlocks"`
	AchvUnlocks        AchvUnlocks        `json:"achvUnlocks"`
	VoucherUnlocks     VoucherUnlocks     `json:"voucherUnlocks"`
	VoucherCounts      VoucherCounts      `json:"voucherCounts"`
	Eggs               []EggData          `json:"eggs"`
	GameVersion        string             `json:"gameVersion"`
	Timestamp          int                `json:"timestamp"`
}

type DexData map[int]DexEntry

type DexEntry struct {
	SeenAttr     interface{} `json:"seenAttr"`   // integer or string
	CaughtAttr   interface{} `json:"caughtAttr"` // integer or string
	NatureAttr   int         `json:"natureAttr"`
	SeenCount    int         `json:"seenCount"`
	CaughtCount  int         `json:"caughtCount"`
	HatchedCount int         `json:"hatchedCount"`
	Ivs          []int       `json:"ivs"`
}

type StarterMoveData map[int]interface{}

type StarterEggMoveData map[int]int

type GameStats interface{}

type Unlocks map[int]bool

type AchvUnlocks map[string]int

type VoucherUnlocks map[string]int

type VoucherCounts map[string]int

type EggData struct {
	Id         int       `json:"id"`
	GachaType  GachaType `json:"gachaType"`
	HatchWaves int       `json:"hatchWaves"`
	Timestamp  int       `json:"timestamp"`
}

type GachaType int

type SessionSaveData struct {
	Seed           string                   `json:"seed"`
	PlayTime       int                      `json:"playTime"`
	GameMode       GameMode                 `json:"gameMode"`
	Party          []PokemonData            `json:"party"`
	EnemyParty     []PokemonData            `json:"enemyParty"`
	Modifiers      []PersistentModifierData `json:"modifiers"`
	EnemyModifiers []PersistentModifierData `json:"enemyModifiers"`
	Arena          ArenaData                `json:"arena"`
	PokeballCounts PokeballCounts           `json:"pokeballCounts"`
	Money          int                      `json:"money"`
	Score          int                      `json:"score"`
	WaveIndex      int                      `json:"waveIndex"`
	BattleType     BattleType               `json:"battleType"`
	Trainer        TrainerData              `json:"trainer"`
	GameVersion    string                   `json:"gameVersion"`
	Timestamp      int                      `json:"timestamp"`
}

type GameMode int

type PokemonData interface{}

type PersistentModifierData interface{}

type ArenaData interface{}

type PokeballCounts map[string]int

type BattleType int

type TrainerData interface{}

type SessionHistoryData struct {
	Seed        string                   `json:"seed"`
	PlayTime    int                      `json:"playTime"`
	Result      SessionHistoryResult     `json:"sessionHistoryResult"`
	GameMode    GameMode                 `json:"gameMode"`
	Party       []PokemonData            `json:"party"`
	Modifiers   []PersistentModifierData `json:"modifiers"`
	Money       int                      `json:"money"`
	Score       int                      `json:"score"`
	WaveIndex   int                      `json:"waveIndex"`
	BattleType  BattleType               `json:"battleType"`
	GameVersion string                   `json:"gameVersion"`
	Timestamp   int                      `json:"timestamp"`
}

type SessionHistoryResult int
