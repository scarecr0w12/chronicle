package dbstatic

import (
	"github.com/google/uuid"
)

func ServerTurtleWoW() uuid.UUID {
	return uuid.MustParse("10ac9e23-ff74-43ed-83ad-96c123017097")
}
func ServerSATurtleWoW() uuid.UUID   { return uuid.MustParse("eaa7e20e-ae86-4690-98e0-dde0b9f06cd0") }
func ServerAsiaTurtleWoW() uuid.UUID { return uuid.MustParse("9750514d-be08-4700-bce7-4108916b7ea0") }
func ServerEpoch() uuid.UUID         { return uuid.MustParse("2f7e2ccc-9aa2-4b48-8ee9-b146a9138d06") }
func ServerUnknown() uuid.UUID {
	return uuid.MustParse("89b9a047-71c7-4f0d-96a0-247308a81f90")
}

func RealmAmbershire() uuid.UUID {
	return uuid.MustParse("851d2fd3-f9c5-4623-b714-924b59d916aa")
}

func RealmTelAbim() uuid.UUID {
	return uuid.MustParse("f94d3103-1cd8-40e9-ad91-a2366de33354")
}

func RealmNordanaar() uuid.UUID {
	return uuid.MustParse("bcf173a7-c94a-49fe-8930-27435d722fb7")
}

// SA

func RealmSouthSeas() uuid.UUID {
	return uuid.MustParse("ad486d39-31dd-4eb6-a43d-7d469df4ffcf")
}

// Asia

func RealmGehennas() uuid.UUID {
	return uuid.MustParse("c240e1e4-9d2b-46f7-b23c-6b55a37b4710")
}

func RealmRavenstorm() uuid.UUID {
	return uuid.MustParse("885cd224-aa71-4592-81e2-98fe138ca650")
}

func RealmKarazhan() uuid.UUID {
	return uuid.MustParse("0f9825e5-8a88-4bfb-80f6-26b472c7a1aa")
}

func RealmBloodRing() uuid.UUID {
	return uuid.MustParse("5f786828-1c60-4360-8b0f-14b7b494be3a")
}

func RealmUnknown() uuid.UUID {
	return uuid.MustParse("f6fb8310-9464-4cf1-a143-aba34f1c3037")
}

// Epoch

func RealmGurubashi() uuid.UUID {
	return uuid.MustParse("e9c0f97b-0b2e-4f47-848c-68634ba6a3dd")
}

func RealmKezan() uuid.UUID {
	return uuid.MustParse("140eaa55-317d-4299-8756-83f495efba15")
}
