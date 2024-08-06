package helpers

// Inspired from determineClass function here: https://github.com/willsunnn/ms-tracker/blob/8b00613ff22a1cea8f93f5081d92cd54284fb4a8/library/src/api/AdditionalCharacterInfoFirebaseApi.ts#L120

func DetermineClass(jobName string, jobDetail int) string {
	switch jobName {
	case "Warrior":
		{
			switch jobDetail {
			case 12:
				return "Hero"
			case 22:
				return "Paladin"
			case 32:
				return "Dark Knight"
			default:
				return "Unknown Warrior"
			}
		}
	case "Magician":
		{
			switch jobDetail {
			case 12:
				return "Fire/Poison Archmage"
			case 22:
				return "Ice/Lightning Archmage"
			case 32:
				return "Bishop"
			default:
				return "Unknown Magician"
			}
		}
	case "Thief":
		{
			switch jobDetail {
			case 12:
				return "Night Lord"
			case 22:
				return "Shadower"
			default:
				return "Unknown Thief"
			}
		}
	case "Dual Blade":
		{
			return "Blade Master"
		}
	case "Bowman":
		{
			switch jobDetail {
			case 12:
				return "Bowmaster"
			case 22:
				return "Marksman"
			default:
				return "Unknown Bowman"
			}
		}
	case "Pirate":
		{
			switch jobDetail {
			case 12:
				return "Buccaneer"
			case 22:
				return "Corsair"
			case 32:
				return "Cannon Master"
			default:
				return "Unknown Pirate"
			}
		}
	case "Pathfinder":
	case "Dawn Warrior":
	case "Blaze Wizard":
	case "Wind Archer":
	case "Night Walker":
	case "Thunder Breaker":
	case "Mihile":
	case "Demon Slayer":
	case "Battle Mage":
	case "Wild Hunter":
	case "Mechanic":
	case "Xenon":
	case "Demon Avenger":
	case "Blaster":
	case "Aran":
	case "Evan":
	case "Mercedes":
	case "Phantom":
	case "Luminous":
	case "Shade":
	case "Kaiser":
	case "Angelic Buster":
	case "Cadena":
	case "Kain":
	case "Illium":
	case "Ark":
	case "Adele":
	case "Khali":
	case "Hayato":
	case "Kanna":
	case "Lara":
	case "Hoyoung":
	case "Zero":
	case "Kinesis":
	case "Lynn":
		return jobName
	default:
		return "Unknown " + jobName
	}
	return "What the fricc are you"
}
