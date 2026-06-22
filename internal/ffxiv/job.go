package ffxiv

type Job int

const (
	GNB Job = iota
	PLD
	GLD
	DRK
	WAR
	MRD
	SCH
	ACN // Arcanist
	SGE
	AST
	WHM
	CNJ
	SAM
	DRG
	NIN
	MNK
	RPR
	VPR
	BRD
	MCH
	DNC
	BLM
	BLU
	SMN
	PCT
	RDM
	LNC
	PUG
	ROG
	THM
	ARC // Archer
	Unknown
)

func JobFromAbbreviation(abbreviation string) Job {
	switch abbreviation {
	case "ACN":
		return ACN
	case "ARC":
		return ARC
	case "AST":
		return AST
	case "BLM":
		return BLM
	case "BLU":
		return BLU
	case "BRD":
		return BRD
	case "CNJ":
		return CNJ
	case "DNC":
		return DNC
	case "DRG":
		return DRG
	case "DRK":
		return DRK
	case "GLD":
		return GLD
	case "GNB":
		return GNB
	case "LNC":
		return LNC
	case "MCH":
		return MCH
	case "MNK":
		return MNK
	case "MRD":
		return MRD
	case "NIN":
		return NIN
	case "PCT":
		return PCT
	case "PLD":
		return PLD
	case "PUG":
		return PUG
	case "RDM":
		return RDM
	case "ROG":
		return ROG
	case "RPR":
		return RPR
	case "SAM":
		return SAM
	case "SCH":
		return SCH
	case "SGE":
		return SGE
	case "SMN":
		return SMN
	case "THM":
		return THM
	case "VPR":
		return VPR
	case "WAR":
		return WAR
	case "WHM":
		return WHM
	}
	return Unknown
}

func (j Job) Emoji() string {
	switch j {
	case ACN:
		return "<:arcanist:1518612053315682509>"
	case ARC:
		return "<:archer:1518612061217882222>"
	case AST:
		return "<:astrologian:1518605444044816447>"
	case BLM:
		return "<:blackmage:1518605442849181768>"
	case BLU:
		return "<:bluemager:1518605465444155412>"
	case BRD:
		return "<:bard:1518605456807821342>"
	case CNJ:
		return "<:conjurer:1518612064594166016>"
	case DNC:
		return "<:dancer:1518605471194288230>"
	case DRG:
		return "<:dragoon:1518605468702871563>"
	case DRK:
		return "<:darkknight:1518605466509250701>"
	case GLD:
		return "<:gladiator:1518612056444764375>"
	case GNB:
		return "<:gunbreaker:1518605453284741222>"
	case LNC:
		return "<:lancer:1518612058067959908>"
	case MCH:
		return "<:machinist:1518605467608416396>"
	case MNK:
		return "<:monk:1518605454748549120>"
	case MRD:
		return "<:marauder:1518612059762200777>"
	case NIN:
		return "<:ninja:1518605457755869316>"
	case PCT:
		return "<:pictomancer:1518605446666129478>"
	case PLD:
		return "<:paladin:1518605461891579977>"
	case PUG:
		return "<:pugilist:1518612062773969068>"
	case RDM:
		return "<:redmage:1518605459806748774>"
	case ROG:
		return "<:rogue:1518612051646353498>"
	case RPR:
		return "<:reaper:1518605464009576490>"
	case SAM:
		return "<:samurai:1518605445395382435>"
	case SCH:
		return "<:scholar:1518605460729626858>"
	case SGE:
		return "<:sage:1518605451351036036>"
	case SMN:
		return "<:summoner:1518605469827072165>"
	case THM:
		return "<:thaumaturge:1518612055232479254>"
	case VPR:
		return "<:viper:1518605463250407434>"
	case WAR:
		return "<:warrior:1518605447970426880>"
	case WHM:
		return "<:whitemage:1518605441905594428>"
	}
	return "<:tankhealerdps:1518614267375390770>"
}
