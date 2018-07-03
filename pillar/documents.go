package pillar

type storeResponse struct {
	Total int `json:"total_sold"`
}

type blenderIDPP struct {
	Latest   int `json:"latest"`
	Obsolete int `json:"obsolete"`
	Never    int `json:"never"`
}

type blenderIDUsers struct {
	ConfirmedEmailCount   int          `json:"confirmed"`
	UnconfirmedEmailCount int          `json:"unconfirmed"`
	TotalCount            int          `json:"total"`
	PrivacyPolicyAgreed   *blenderIDPP `json:"privacy_policy_agreed" bson:"privacy_policy_agreed"`
}

type blenderIDResponse struct {
	Users blenderIDUsers `json:"users"`
}
