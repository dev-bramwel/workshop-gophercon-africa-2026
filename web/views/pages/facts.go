package pages

import "workshop/internal/domain/species"

func AnimalFacts(a AnimalCard) []Fact {
	source := "https://animals.sandiegozoo.org/animals/"
	switch a.Species.GBIFKey {
	case 5219404:
		return []Fact{{"A voice that travels", "A lion’s roar can be heard up to three miles away. It helps a pride communicate and announce its territory.", source + "lion", "San Diego Zoo Wildlife Alliance"}, {"A social big cat", "Lions live in prides. Related females often stay together and help care for one another’s cubs.", source + "lion", "San Diego Zoo Wildlife Alliance"}}
	case 2435350, 2435349, 5219461:
		return []Fact{{"Ears that keep their cool", "An elephant’s ears help it cool down. Blood flowing through the ears releases heat as the animal flaps them.", source + "elephant", "San Diego Zoo Wildlife Alliance"}, {"A remarkably versatile nose", "An elephant’s trunk combines its nose and upper lip. It is used to breathe, smell, drink and grasp food.", source + "elephant", "San Diego Zoo Wildlife Alliance"}}
	case 2440892, 2440888, 2440894:
		return []Fact{{"A pattern all their own", "Each zebra has its own stripe pattern. Even members of the same herd have different markings.", source + "zebra", "San Diego Zoo Wildlife Alliance"}, {"Stronger together", "Zebras are social animals. Living in groups lets more eyes and ears stay alert for predators.", source + "zebra", "San Diego Zoo Wildlife Alliance"}}
	case 7716694, 2441205, 9163257, 8959277:
		return []Fact{{"Seven, just like us", "A giraffe’s remarkably long neck has seven vertebrae, the same number as a human neck.", source + "giraffe", "San Diego Zoo Wildlife Alliance"}, {"A tongue built for browsing", "A giraffe’s long, grasping tongue helps it pull leaves from branches, even among thorns.", source + "giraffe", "San Diego Zoo Wildlife Alliance"}}
	case 5220086:
		return []Fact{{"Fishing with bubbles", "Some humpback whales coordinate to surround prey with curtains of bubbles, concentrating food before they feed.", "https://www.fisheries.noaa.gov/species/humpback-whale/science", "NOAA Fisheries"}, {"A life near the surface", "Humpbacks breach and slap the water with their tails and long pectoral fins, making them especially recognisable to whale watchers.", "https://www.fisheries.noaa.gov/species/humpback-whale/science", "NOAA Fisheries"}}
	}
	facts := extraAnimalFacts(a.Species)
	if a.Species.Family != "" {
		facts = append(facts, Fact{"A place on the family tree", Name(a.Species) + " belongs to the " + a.Species.Family + " family. Taxonomic families group organisms with shared evolutionary ancestry.", GBIF(a.Species), "GBIF taxonomy"})
	}
	if a.Species.Genus != "" {
		facts = append(facts, Fact{"The name behind the animal", "Its genus is " + a.Species.Genus + ". In a two-part scientific name, the genus comes first and connects closely related species.", GBIF(a.Species), "GBIF taxonomy"})
	}
	if len(facts) == 0 {
		facts = append(facts, Fact{"A record worth exploring", "Every observation adds a place and time to our understanding of this animal. Explore the original taxonomy record for more detail.", GBIF(a.Species), "GBIF taxonomy"})
	}
	return facts[:min(2, len(facts))]
}

// Facts are kept locally, with a direct source on every card; browsing a guide
// never depends on another API being available.
func extraAnimalFacts(s species.Species) []Fact {
	zoo := func(slug, title, text string) []Fact {
		return []Fact{{title, text, "https://animals.sandiegozoo.org/animals/" + slug, "San Diego Zoo Wildlife Alliance"}}
	}
	noaa := func(slug, title, text string) []Fact {
		return []Fact{{title, text, "https://www.fisheries.noaa.gov/species/" + slug, "NOAA Fisheries"}}
	}
	aquarium := func(slug, title, text string) []Fact {
		return []Fact{{title, text, "https://www.montereybayaquarium.org/animals-the-ocean/animals-a-to-z/" + slug, "Monterey Bay Aquarium"}}
	}
	switch s.GBIFKey {
	case 2435270:
		return zoo("cheetah", "A tail that steers", "A cheetah’s long tail helps it steer during a sprint, while its exposed claws provide traction like running spikes.")
	case 5219436:
		return zoo("leopard", "Dinner in the treetops", "Leopards can carry a kill heavier than themselves into a tree, keeping their meal beyond the reach of competing predators.")
	case 5219426:
		return zoo("jaguar", "At home in the water", "Jaguars are strong swimmers. They often live near rivers and streams, where aquatic animals can become part of their diet.")
	case 5219416:
		return zoo("tiger", "Hidden in plain sight", "A tiger’s stripes break up its outline among grasses and trees, helping this ambush hunter approach without being noticed.")
	case 2441247:
		return zoo("hippo", "A built-in skin shield", "A hippo’s reddish skin secretion helps keep its skin moist and protects it from sunburn. Despite its appearance, it is not blood.")
	case 5786666:
		return zoo("capybara", "An underwater escape", "Capybaras can stay underwater for up to five minutes. Their eyes, ears and nostrils sit high on the head, letting them watch while mostly submerged.")
	case 2433376:
		return zoo("platypus", "Finding food by electricity", "A platypus closes its eyes and ears underwater. Receptors in its flexible bill detect tiny electrical signals from living prey.")
	case 2440301:
		return zoo("wombat", "A pouch made for digging", "A female wombat’s pouch opens towards her rear. This keeps loose soil from filling it while she digs.")
	case 2440012:
		return zoo("koala", "The art of taking it slow", "Eucalypt leaves provide little energy. Koalas conserve theirs by resting and sleeping for much of the day.")
	case 2433451:
		return zoo("polar-bear", "An unexpected coat", "A polar bear’s white-looking coat is made of clear, hollow guard hairs over a dense undercoat. Water rolls off the outer hairs.")
	case 4352332:
		return zoo("flamingo", "You are what you eat", "Flamingos get their pink and reddish tones from carotenoid pigments in their food, including algae and small crustaceans.")
	case 5218781:
		return zoo("spotted-hyena", "A society led by females", "Spotted hyenas live in complex, female-dominated clans. Females are generally larger than males and hold the higher social ranks.")
	case 2470854:
		return zoo("komodo-dragon", "Armour beneath the skin", "The world’s largest living lizard has bony plates called osteoderms reinforcing its skin, adding protection against scratches and bites.")
	case 2436412:
		return zoo("lemur", "A battle of scents", "Male ring-tailed lemurs sometimes rub scent from their wrists onto their tails, then wave them at rivals in a so-called stink fight.")
	case 5219532, 5707420, 9311132:
		return zoo("orangutan", "A fresh bed in the canopy", "Orangutans usually build a new sleeping nest in the trees each night. They also use leafy branches to shelter from rain and sun.")
	case 2442225:
		return noaa("green-turtle", "The green is on the inside", "The green turtle takes its name from the colour of its body fat. Its shell can be brown, grey or olive.")
	case 9789983:
		return noaa("leatherback-turtle", "A different kind of shell", "The leatherback is the only sea turtle without a hard shell and scales. Its carapace is covered in tough, rubbery skin.")
	case 2433670:
		return aquarium("sea-otter", "Warmth woven into fur", "Sea otters rely on exceptionally dense fur rather than a thick layer of blubber. Grooming traps insulating air within their coat.")
	case 2420694:
		return aquarium("white-shark", "A hidden electrical sense", "A white shark can sense the faint electrical fields produced by living animals, helping it locate potential prey.")
	case 5229384:
		return zoo("african-penguin", "A face that helps it cool", "The bare skin above an African penguin’s eyes helps release heat. More blood flows through these patches as the bird gets hot.")
	}
	switch s.Family {
	case "Rhinocerotidae":
		return zoo("rhinoceros", "A horn made like your nails", "Rhino horns are made of keratin, the protein found in human hair and fingernails. They have no bony core.")
	case "Alligatoridae", "Crocodylidae", "Gavialidae":
		return zoo("crocodilian", "Listening before hatching", "Crocodilians can hear their young calling from inside their eggs. Many species guard their nests and protect their hatchlings.")
	case "Balaenopteridae", "Balaenidae", "Eschrichtiidae":
		return []Fact{{"A filter made of keratin", "Baleen whales strain food through plates made of keratin, the same material as human fingernails. They have two blowholes instead of one.", "https://oceantoday.noaa.gov/whaleanatomy/", "NOAA Ocean Today"}}
	case "Delphinidae", "Monodontidae", "Physeteridae", "Iniidae":
		return []Fact{{"Seeing with sound", "Toothed whales, including dolphins, use echolocation clicks. Returning echoes help them navigate and find food in the water.", "https://www.fisheries.noaa.gov/new-england-mid-atlantic/science-data/marine-mammal-acoustics", "NOAA Fisheries"}}
	}
	return nil
}
