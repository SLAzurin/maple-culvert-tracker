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
				return "Markman" // Fight me. It's Markman not Marksman.
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
		fallthrough
	case "Dawn Warrior":
		fallthrough
	case "Blaze Wizard":
		fallthrough
	case "Wind Archer":
		fallthrough
	case "Night Walker":
		fallthrough
	case "Thunder Breaker":
		fallthrough
	case "Mihile":
		fallthrough
	case "Demon Slayer":
		fallthrough
	case "Battle Mage":
		fallthrough
	case "Wild Hunter":
		fallthrough
	case "Mechanic":
		fallthrough
	case "Xenon":
		fallthrough
	case "Demon Avenger":
		fallthrough
	case "Blaster":
		fallthrough
	case "Aran":
		fallthrough
	case "Evan":
		fallthrough
	case "Mercedes":
		fallthrough
	case "Phantom":
		fallthrough
	case "Luminous":
		fallthrough
	case "Shade":
		fallthrough
	case "Kaiser":
		fallthrough
	case "Angelic Buster":
		fallthrough
	case "Cadena":
		fallthrough
	case "Kain":
		fallthrough
	case "Illium":
		fallthrough
	case "Ark":
		fallthrough
	case "Adele":
		fallthrough
	case "Khali":
		fallthrough
	case "Hayato":
		fallthrough
	case "Kanna":
		fallthrough
	case "Lara":
		fallthrough
	case "Hoyoung":
		fallthrough
	case "Zero":
		fallthrough
	case "Kinesis":
		fallthrough
	case "Lynn":
		return jobName
	default:
		return "Unknown " + jobName
	}
}
