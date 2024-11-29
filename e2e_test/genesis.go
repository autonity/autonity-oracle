package test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"math/big"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config     *ChainConfig   `json:"config"`
	Nonce      uint64         `json:"nonce"`
	Timestamp  uint64         `json:"timestamp"`
	ExtraData  []byte         `json:"extraData"`
	GasLimit   uint64         `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int       `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash    `json:"mixHash"`
	Coinbase   common.Address `json:"coinbase"`
	Alloc      GenesisAlloc   `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`

	BaseFee *big.Int `json:"baseFeePerGas"`
}

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code          []byte                      `json:"code,omitempty"`
	Storage       map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance       *big.Int                    `json:"balance" gencodec:"required"`
	NewtonBalance *big.Int                    `json:"newtonBalance"`
	// validator address to amount bond to this validator
	Bonds      map[common.Address]*big.Int `json:"bonds"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	BaseFee    *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code          hexutil.Bytes
	Balance       *math.HexOrDecimal256
	NewtonBalance *math.HexOrDecimal256
	Bonds         map[common.Address]*math.HexOrDecimal256
	Nonce         math.HexOrDecimal64
	Storage       map[storageJSON]storageJSON
	PrivateKey    hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// types to construct genesis.

type AutonityContractGenesis struct {
	Bytecode                hexutil.Bytes         `json:"bytecode,omitempty" toml:",omitempty"`
	ABI                     *abi.ABI              `json:"abi,omitempty" toml:",omitempty"`
	MinBaseFee              uint64                `json:"minBaseFee"`
	EpochPeriod             uint64                `json:"epochPeriod"`
	UnbondingPeriod         uint64                `json:"unbondingPeriod"`
	BlockPeriod             uint64                `json:"blockPeriod"`
	MaxCommitteeSize        uint64                `json:"maxCommitteeSize"`
	MaxScheduleDuration     uint64                `json:"maxScheduleDuration"`
	Operator                common.Address        `json:"operator"`
	Treasury                common.Address        `json:"treasury"`
	WithheldRewardsPool     common.Address        `json:"withheldRewardsPool"`
	TreasuryFee             uint64                `json:"treasuryFee"`
	DelegationRate          uint64                `json:"delegationRate"`
	WithholdingThreshold    uint64                `json:"withholdingThreshold"`
	ProposerRewardRate      uint64                `json:"proposerRewardRate"`
	OracleRewardRate        uint64                `json:"oracleRewardRate"`
	InitialInflationReserve *math.HexOrDecimal256 `json:"initialInflationReserve"`
	Validators              []*Validator          `json:"validators"` // todo: Can we change that to []Validator
	Schedules               []Schedule            `json:"schedules"`
}

type AccountabilityGenesis struct {
	InnocenceProofSubmissionWindow uint64 `json:"innocenceProofSubmissionWindow"`

	// Slashing parameters
	BaseSlashingRateLow  uint64 `json:"baseSlashingRateLow"`
	BaseSlashingRateMid  uint64 `json:"baseSlashingRateMid"`
	BaseSlashingRateHigh uint64 `json:"baseSlashingRateHigh"`

	// Factors
	CollusionFactor uint64 `json:"collusionFactor"`
	HistoryFactor   uint64 `json:"historyFactor"`
	JailFactor      uint64 `json:"jailFactor"`
}

// OmissionAccountabilityGenesis defines the omission fault detection parameters
type OmissionAccountabilityGenesis struct {
	InactivityThreshold    uint64 `json:"inactivityThreshold"`
	LookbackWindow         uint64 `json:"LookbackWindow"`
	PastPerformanceWeight  uint64 `json:"pastPerformanceWeight"` // k belong to [0, 1), after scaling in the contract
	InitialJailingPeriod   uint64 `json:"initialJailingPeriod"`
	InitialProbationPeriod uint64 `json:"initialProbationPeriod"`
	InitialSlashingRate    uint64 `json:"initialSlashingRate"`
	Delta                  uint64 `json:"delta"`
}

// OracleContractGenesis Autonity contract config. It is used for deployment.
type OracleContractGenesis struct {
	Bytecode                  hexutil.Bytes `json:"bytecode,omitempty" toml:",omitempty"`
	ABI                       *abi.ABI      `json:"abi,omitempty" toml:",omitempty"`
	Symbols                   []string      `json:"symbols"`
	VotePeriod                uint64        `json:"votePeriod"`
	OutlierDetectionThreshold uint64        `json:"outlierDetectionThreshold"`
	OutlierSlashingThreshold  uint64        `json:"outlierSlashingThreshold"`
	BaseSlashingRate          uint64        `json:"baseSlashingRate"`
}

type AsmConfig struct {
	ACUContractConfig           *AcuContractGenesis           `json:"acu,omitempty"`
	StabilizationContractConfig *StabilizationContractGenesis `json:"stabilization,omitempty"`
	SupplyControlConfig         *SupplyControlGenesis         `json:"supplyControl,omitempty"`
}

type AcuContractGenesis struct {
	Symbols    []string
	Quantities []uint64
	Scale      uint64
}

type StabilizationContractGenesis struct {
	BorrowInterestRate        *math.HexOrDecimal256
	LiquidationRatio          *math.HexOrDecimal256
	MinCollateralizationRatio *math.HexOrDecimal256
	MinDebtRequirement        *math.HexOrDecimal256
	TargetPrice               *math.HexOrDecimal256
}

type SupplyControlGenesis struct {
	InitialAllocation *math.HexOrDecimal256
}

type InflationControllerGenesis struct {
	// Those parameters need to be compatible with the solidity SD59x18 format
	InflationRateInitial      *math.HexOrDecimal256 `json:"inflationRateInitial"`
	InflationRateTransition   *math.HexOrDecimal256 `json:"inflationRateTransition"`
	InflationReserveDecayRate *math.HexOrDecimal256 `json:"inflationReserveDecayRate"`
	InflationTransitionPeriod *math.HexOrDecimal256 `json:"inflationTransitionPeriod"`
	InflationCurveConvexity   *math.HexOrDecimal256 `json:"inflationCurveConvexity"`
}

type NonStakeableVestingGenesis struct {
	NonStakeableContracts []NonStakeableVestingData `json:"nonStakeableVestingContracts"`
}

type Schedule struct {
	Start         *big.Int       `json:"startTime"`
	TotalDuration *big.Int       `json:"totalDuration"`
	Amount        *big.Int       `json:"amount"`
	VaultAddress  common.Address `json:"vaultAddress"`
}

type NonStakeableVestingData struct {
	Beneficiary   common.Address `json:"beneficiary"`
	Amount        *big.Int       `json:"amount"`
	ScheduleID    *big.Int       `json:"scheduleID"`
	CliffDuration *big.Int       `json:"cliffDuration"`
}

type StakeableVestingGenesis struct {
	TotalNominal       *big.Int               `json:"totalNominal"`
	StakeableContracts []StakeableVestingData `json:"stakeableVestingContracts"`
}

type StakeableVestingData struct {
	Beneficiary   common.Address `json:"beneficiary"`
	Amount        *big.Int       `json:"amount"`
	Start         *big.Int       `json:"startTime"`
	CliffDuration *big.Int       `json:"cliffDuration"`
	TotalDuration *big.Int       `json:"totalDuration"`
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int    `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP150Hash  common.Hash `json:"eip150Hash,omitempty"`  // EIP150 HF hash (needed for header only clients as only gas pricing changed)

	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as Constantinople)
	IstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // Istanbul switch block (nil = no fork, 0 = already on istanbul)
	MuirGlacierBlock    *big.Int `json:"muirGlacierBlock,omitempty"`    // Eip-2384 (bomb delay) switch block (nil = no fork, 0 = already activated)
	BerlinBlock         *big.Int `json:"berlinBlock,omitempty"`         // Berlin switch block (nil = no fork, 0 = already on berlin)
	LondonBlock         *big.Int `json:"londonBlock,omitempty"`         // London switch block (nil = no fork, 0 = already on london)
	ArrowGlacierBlock   *big.Int `json:"arrowGlacierBlock,omitempty"`   // Eip-4345 (bomb delay) switch block (nil = no fork, 0 = already activated)
	MergeForkBlock      *big.Int `json:"mergeForkBlock,omitempty"`      // EIP-3675 (TheMerge) switch block (nil = no fork, 0 = already in merge proceedings)

	// TerminalTotalDifficulty is the amount of total difficulty reached by
	// the network that triggers the consensus upgrade.
	TerminalTotalDifficulty *big.Int `json:"terminalTotalDifficulty,omitempty"`

	// Various consensus engines
	Ethash                       *EthashConfig                  `json:"ethash,omitempty"`
	AutonityContractConfig       *AutonityContractGenesis       `json:"autonity,omitempty"`
	AccountabilityConfig         *AccountabilityGenesis         `json:"accountability,omitempty"`
	OracleContractConfig         *OracleContractGenesis         `json:"oracle,omitempty"`
	InflationContractConfig      *InflationControllerGenesis    `json:"inflation,omitempty"`
	ASM                          AsmConfig                      `json:"asm,omitempty"`
	NonStakeableVestingConfig    *NonStakeableVestingGenesis    `json:"nonStakeableVesting,omitempty"`
	StakeableVestingConfig       *StakeableVestingGenesis       `json:"stakeableVesting,omitempty"`
	OmissionAccountabilityConfig *OmissionAccountabilityGenesis `json:"omissionAccountability,omitempty"`

	// true if run in testmode, false by default
	TestMode bool `json:"testMode,omitempty"`
}

type Validator struct {
	Treasury      common.Address `json:"treasury"`
	Enode         string         `json:"enode"`
	OracleAddress common.Address `json:"oracleAddress"`
	BondedStake   *big.Int       `json:"bondedStake"`
	ConsensusKey  string         `json:"consensusKey"`
}
