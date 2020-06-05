package timeline

import "github.com/idirall22/twee/pb"

// Server timeline server service
type Server struct {
}

// NewTimelineServer create new timeline server
func NewTimelineServer() (*Server, error) {
	return nil, nil
}

// List timeline
func (s *Server) List(req *pb.TimelineRequest, stream pb.TimelineService_ListServer) error {
	return nil
}
