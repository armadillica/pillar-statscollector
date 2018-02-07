package pillar

type storeResponse struct {
	Total int `json:"total_sold"`
}

type blenderIDUsers struct {
	ConfirmedEmailCount   int `json:"confirmed"`
	UnconfirmedEmailCount int `json:"unconfirmed"`
	TotalCount            int `json:"total"`
}

type blenderIDResponse struct {
	Users blenderIDUsers `json:"users"`
}
