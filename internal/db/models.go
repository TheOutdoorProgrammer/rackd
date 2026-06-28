package db

// Every model below is stored as a single encrypted JSON blob per row. The ID
// (and any foreign key) is the plaintext column; all other fields are sealed.

// Firearm is a single firearm or NFA item.
type Firearm struct {
	ID                 int64    `json:"id"`
	Nickname           string   `json:"nickname"`
	Manufacturer       string   `json:"manufacturer"`
	Model              string   `json:"model"`
	Kind               string   `json:"kind"` // pistol | rifle | shotgun | nfa | other
	Caliber            string   `json:"caliber"`
	ShellLengths       []string `json:"shellLengths"` // shotgun chamber: supported shell lengths
	SerialNumber       string   `json:"serialNumber"`
	Finish             string   `json:"finish"`
	AcquiredDate       string   `json:"acquiredDate"`
	AcquiredPriceCents int64    `json:"acquiredPriceCents"`
	AcquiredFrom       string   `json:"acquiredFrom"`
	Status             string   `json:"status"` // owned | sold | loaned | pending
	IsNFA              bool     `json:"isNfa"`
	NFAType            string   `json:"nfaType"` // suppressor | sbr | sbs | mg | aow
	TaxStampDate       string   `json:"taxStampDate"`
	Notes              string   `json:"notes"`
	CreatedAt          string   `json:"createdAt"`
	UpdatedAt          string   `json:"updatedAt"`
}

// Ammo is a line of ammunition stock.
type Ammo struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Caliber            string `json:"caliber"`
	Brand              string `json:"brand"`
	BulletType         string `json:"bulletType"`  // FMJ | JHP | HP | SP | match | birdshot | buckshot | slug | other
	ShellLength        string `json:"shellLength"` // shotshells: 2½" | 2¾" | 3" | 3½"
	GrainWeight        int64  `json:"grainWeight"`
	QuantityOnHand     int64  `json:"quantityOnHand"`
	LotNumber          string `json:"lotNumber"`
	AcquiredDate       string `json:"acquiredDate"`
	AcquiredPriceCents int64  `json:"acquiredPriceCents"`
	AcquiredFrom       string `json:"acquiredFrom"`
	Notes              string `json:"notes"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

// Knife is a single knife.
type Knife struct {
	ID                 int64  `json:"id"`
	Nickname           string `json:"nickname"`
	Type               string `json:"type"` // folding | fixed | automatic | balisong | multitool | machete | other
	Manufacturer       string `json:"manufacturer"`
	Model              string `json:"model"`
	BladeSteel         string `json:"bladeSteel"`
	BladeLengthIn      string `json:"bladeLengthIn"`
	SerialNumber       string `json:"serialNumber"`
	AcquiredDate       string `json:"acquiredDate"`
	AcquiredPriceCents int64  `json:"acquiredPriceCents"`
	AcquiredFrom       string `json:"acquiredFrom"`
	LastSharpenedDate  string `json:"lastSharpenedDate"`
	Status             string `json:"status"`
	Notes              string `json:"notes"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

// Accessory optionally hangs off a firearm (firearm_id is a plaintext column so
// the relationship survives encryption).
type Accessory struct {
	ID           int64  `json:"id"`
	FirearmID    *int64 `json:"firearmId"`
	Name         string `json:"name"`
	Category     string `json:"category"` // optic | light | laser | sling | magazine | trigger | stock | case | cleaning | other
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	SerialNumber string `json:"serialNumber"`
	ValueCents   int64  `json:"valueCents"`
	Quantity     int64  `json:"quantity"`
	AcquiredFrom string `json:"acquiredFrom"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// AmmoLink is an ammo line associated with a firearm, plus an optional note
// (e.g. "zeroed / preferred load").
type AmmoLink struct {
	Ammo Ammo   `json:"ammo"`
	Note string `json:"note"`
}

// Attachment is an uploaded photo. The file bytes live encrypted on disk under
// StoredPath/ThumbPath (opaque names, plaintext columns); descriptive metadata
// is sealed in the row's data blob.
type Attachment struct {
	ID          int64  `json:"id"`
	OwnerType   string `json:"ownerType"`
	OwnerID     int64  `json:"ownerId"`
	Kind        string `json:"kind"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	CreatedAt   string `json:"createdAt"`
	Cover       bool   `json:"cover"` // shown on list cards
	StoredPath  string `json:"-"`
	ThumbPath   string `json:"-"`
}
