package api

//ActivityMember exists to essentially remove the XXX_ tags from protobuf generated structs
// I'm not sure if this is the best approach; I'm not a huge fan of it
// But I also don't want tags like that clogging up my database and I don't know a better solution
type ActivityMember struct {
	Name       string            `json:"name" bson:"name"`
	Activities map[string]uint32 `json:"activities" bson:"activities"`
}
