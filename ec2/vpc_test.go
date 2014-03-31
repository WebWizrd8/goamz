package ec2_test

import (
	"github.com/goamz/goamz/ec2"
	"github.com/motain/gocheck"
)

func (s *S) TestDescribeRouteTables(c *gocheck.C) {
	testServer.Response(200, nil, DescribeRouteTablesExample)

	filter := ec2.NewFilter()
	filter.Add("key1", "value1")
	filter.Add("key2", "value2", "value3")

	resp, err := s.ec2.DescribeRouteTables([]string{"rt1", "rt2"}, nil)

	req := testServer.WaitRequest()
	c.Assert(req.Form["Action"], gocheck.DeepEquals, []string{"DescribeRouteTables"})
	c.Assert(req.Form["RouteTableId.1"], gocheck.DeepEquals, []string{"rt1"})
	c.Assert(req.Form["RouteTableId.2"], gocheck.DeepEquals, []string{"rt2"})

	c.Assert(err, gocheck.IsNil)
	c.Assert(resp.RequestId, gocheck.Equals, "6f570b0b-9c18-4b07-bdec-73740dcf861aEXAMPLE")
	c.Assert(resp.RouteTables, gocheck.HasLen, 2)

	rt1 := resp.RouteTables[0]
	c.Assert(rt1.Id, gocheck.Equals, "rtb-13ad487a")
	c.Assert(rt1.VpcId, gocheck.Equals, "vpc-11ad4878")
	c.Assert(rt1.Routes, gocheck.DeepEquals, []ec2.Route{
		{DestinationCidrBlock: "10.0.0.0/22", GatewayId: "local", State: "active", Origin: "CreateRouteTable"},
	})
	c.Assert(rt1.Associations, gocheck.DeepEquals, []ec2.RouteTableAssociation{
		{Id: "rtbassoc-12ad487b", RouteTableId: "rtb-13ad487a", Main: true},
	})

	rt2 := resp.RouteTables[1]
	c.Assert(rt2.Id, gocheck.Equals, "rtb-f9ad4890")
	c.Assert(rt2.VpcId, gocheck.Equals, "vpc-11ad4878")
	c.Assert(rt2.Routes, gocheck.DeepEquals, []ec2.Route{
		{DestinationCidrBlock: "10.0.0.0/22", GatewayId: "local", State: "active", Origin: "CreateRouteTable"},
		{DestinationCidrBlock: "0.0.0.0/0", GatewayId: "igw-eaad4883", State: "active"},
	})
	c.Assert(rt2.Associations, gocheck.DeepEquals, []ec2.RouteTableAssociation{
		{Id: "rtbassoc-faad4893", RouteTableId: "rtb-f9ad4890", SubnetId: "subnet-15ad487c"},
	})
}
