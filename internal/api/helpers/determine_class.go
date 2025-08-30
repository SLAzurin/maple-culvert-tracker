package helpers

func DetermineClass(jobID int, jobDetail int) string {
	switch jobID {
	case 0:
		return "Beginner"
	case 1:
		{
			switch jobDetail {
			case 12:
				return "Hero"
			case 22:
				return "Paladin"
			case 32:
				return "Dark Knight"
			}
		}
		return "Explorer Warrior"
	case 2:
		{
			switch jobDetail {
			case 12:
				return "Fire/Poison Archmage"
			case 22:
				return "Ice/Lightning Archmage"
			case 32:
				return "Bishop"
			}
		}
		return "Explorer Magician"
	case 3:
		{
			switch jobDetail {
			case 12:
				return "Bowmaster"
			case 22:
				return "Markman" // Fight me. It's Markman not Marksman.
			case 32:
				return "Pathfinder"
			}
		}
		return "Explorer Bowman"
	case 4:
		{
			switch jobDetail {
			case 12:
				return "Night Lord"
			case 22:
				return "Shadower"
			case 34:
				return "Blade Master"
			}
		}
		return "Explorer Thief"
	case 5:
		{
			switch jobDetail {
			case 12:
				return "Buccaneer"
			case 22:
				return "Corsair"
			case 32:
				return "Cannon Master"
			}
		}
		return "Explorer Pirate"
	case 10:
		return "Noblesse"
	case 11:
		return "Dawn Warrior"
	case 12:
		return "Blaze Wizard"
	case 13:
		return "Wind Archer"
	case 14:
		return "Night Walker"
	case 15:
		return "Thunder Breaker"
	case 202:
		return "Mihile"
	case 30:
		return "Citizen"
	case 31:
		return "Demon Slayer"
	case 32:
		return "Battle Mage"
	case 33:
		return "Wild Hunter"
	case 35:
		return "Mechanic"
	case 208:
		return "Xenon"
	case 209:
		return "Demon Avenger"
	case 215:
		return "Blaster"
	case 20:
		return "Legend"
	case 21:
		return "Aran"
	case 22:
		return "Evan"
	case 23:
		return "Mercedes"
	case 24:
		return "Phantom"
	case 203:
		return "Luminous"
	case 212:
		return "Shade"
	case 204:
		return "Kaiser"
	case 205:
		return "Angelic Buster"
	case 216:
		return "Cadena"
	case 222:
		return "Kain"
	case 217:
		return "Illium"
	case 218:
		return "Ark"
	case 221:
		return "Adele"
	case 224:
		return "Khali"
	case 206:
		return "Hayato"
	case 207:
		return "Kanna"
	case 223:
		return "Lara"
	case 220:
		return "Hoyoung"
	case 210:
		return "Zero"
	case 214:
		return "Kinesis"
	case 225:
		return "Lynn"
	case 226:
		return "Mo Xuan"
	case 227:
		return "Sia Astelle"
	default:
		return "Unknown"
	}
}
