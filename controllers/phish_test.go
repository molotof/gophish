package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gophish/gophish/models"
)

func (s *ControllersSuite) getFirstCampaign() models.Campaign {
	campaigns, err := models.GetCampaigns(1)
	s.Nil(err)
	return campaigns[0]
}

func (s *ControllersSuite) openEmail(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/track?rid=%s", ps.URL, rid))
	s.Nil(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	expected, err := ioutil.ReadFile("static/images/pixel.png")
	s.Nil(err)
	s.Equal(bytes.Compare(body, expected), 0)
}

func (s *ControllersSuite) openEmail404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/track?rid=%s", ps.URL, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) clickLink(rid string, campaign models.Campaign) {
	resp, err := http.Get(fmt.Sprintf("%s/?rid=%s", ps.URL, rid))
	s.Nil(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	s.Equal(bytes.Compare(body, []byte(campaign.Page.HTML)), 0)
}

func (s *ControllersSuite) clickLink404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/?rid=%s", ps.URL, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) TestOpenedPhishingEmail() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.STATUS_SENDING)

	s.openEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EVENT_OPENED)
}

func (s *ControllersSuite) TestClickedPhishingLinkAfterOpen() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.STATUS_SENDING)

	s.openEmail(result.RId)
	s.clickLink(result.RId, campaign)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EVENT_CLICKED)
}

func (s *ControllersSuite) TestNoRecipientID() {
	resp, err := http.Get(fmt.Sprintf("%s/track", ps.URL))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)

	resp, err = http.Get(ps.URL)
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) TestInvalidRecipientID() {
	rid := "XXXXXXXXXX"
	resp, err := http.Get(fmt.Sprintf("%s/track?rid=%s", ps.URL, rid))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)

	resp, err = http.Get(fmt.Sprintf("%s/?rid=%s", ps.URL, rid))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) TestCompletedCampaignClick() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.STATUS_SENDING)
	s.openEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EVENT_OPENED)

	models.CompleteCampaign(campaign.Id, 1)
	s.openEmail404(result.RId)
	s.clickLink404(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EVENT_OPENED)
}

func (s *ControllersSuite) TestRobotsHandler() {
	expected := []byte("User-agent: *\nDisallow: /\n")
	resp, err := http.Get(fmt.Sprintf("%s/robots.txt", ps.URL))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusOK)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	s.Equal(bytes.Compare(body, expected), 0)
}
