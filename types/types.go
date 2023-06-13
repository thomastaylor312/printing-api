package types

type ShippingMethod string
type PaperFinish string

const (
	// TODO: Actually specify methods
	ShippingMethodStandard  ShippingMethod = "standard"
	ShippingMethodExpress   ShippingMethod = "express"
	ShippingMethodOvernight ShippingMethod = "overnight"
	PaperFinishGlossy       PaperFinish    = "glossy"
	PaperFinishMatte        PaperFinish    = "matte"
	PaperFinishLuster       PaperFinish    = "luster"
)

// User represents a user in the system
type User struct {
	UserId   uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
}

func (u *User) ID() uint {
	return u.UserId
}

func (u *User) SetID(id uint) {
	u.UserId = id
}

// Print represents a print that a user wants to order
type Print struct {
	PrintID     uint    `json:"id"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	BorderSize  float64 `json:"borderSize"`
	PaperTypeID uint    `json:"paperTypeId"`
	PictureUrl  string  `json:"pictureUrl"`
	CropX       *uint   `json:"cropX"`
	CropY       *uint   `json:"cropY"`
	Cost        float64 `json:"cost"`
	Quantity    uint    `json:"quantity"`
}

func (p *Print) ID() uint {
	return p.PrintID
}

func (p *Print) SetID(id uint) {
	p.PrintID = id
}

// ShippingDetails represents the shipping details for an order
type ShippingDetails struct {
	Address        string         `json:"address"`
	ShippingCost   float64        `json:"shippingCost"`
	ShippingMethod ShippingMethod `json:"shippingMethod"`
	TrackingNumber *string        `json:"trackingNumber"`
}

// Order represents an order in the system
type Order struct {
	OrderID         uint            `json:"id"`
	UserID          uint            `json:"userId"`
	Prints          []Print         `json:"prints"`
	ShippingDetails ShippingDetails `json:"shippingDetails"`
	HasShipped      bool            `json:"hasShipped"`
	IsDelivered     bool            `json:"isDelivered"`
}

func (o *Order) ID() uint {
	return o.OrderID
}

func (o *Order) SetID(id uint) {
	o.OrderID = id
}

// Cart represents the current items in a user's cart
type Cart struct {
	UserID uint    `json:"userId"`
	Prints []Print `json:"prints"`
}

// PaperType represents a type of paper that can be used for printing and its cost
type PaperType struct {
	PaperID           uint        `json:"id"`
	Name              string      `json:"name"`
	CostPerSquareInch float64     `json:"costPerSquareInch"`
	Finish            PaperFinish `json:"finish"`
}

func (p *PaperType) ID() uint {
	return p.PaperID
}

func (p *PaperType) SetID(id uint) {
	p.PaperID = id
}

// SupplyCosts represents the costs of supplies for printing
type SupplyCosts struct {
	InkPerSquareInch             float64 `json:"inkPerSquareInch"`
	AdditionalSupplyCostPerPrint float64 `json:"additionalSupplyCostPerPrint"`
	DesiredProfitMargin          float64 `json:"desiredProfitMargin"`
}

type Config struct {
	// The max size of the image on its shortest side in inches
	MaxSize float64     `json:"maxSize"`
	Costs   SupplyCosts `json:"costs"`
}
