package endpoint

import (
	"errors"
	"sort"
)

var (
	// AuthenticationFailed is when no authenticator could authenticate an agent
	AuthenticationFailed = errors.New("Authentication failed")
)

// Authenticator is the general interface for authenticators; should return OK
// if the agent is authenticated, DECLINED if this authenticator can't
// authenticate the agent and the auth information should be passed on to next
// authenticator, or FAIL if there's an error and should stop without passing
// on;
//
// The authenticator should respond with Result IF AND ONLY IF the agent is
// authenticated by this authenticator. If hub receives OK, it assumes an
// authenticator has already responded the agent; If hub receives DECLINED, it
// assumes no authenticator can authenticate this agent and nothing has been
// responded, thus will respond with an Error. To make it clearer, no Error
// should be responded by an authenticator -- either respond with an
// authenticated Result, or don't respond at all.
type Authenticator interface {
	Authenticate(agentName string, agentID string, token string, responder *Responder, connCtx ConnContext) HandleCode
}

type authenticatorListItem struct {
	authenticator Authenticator

	// the lower the number, the higher the priority, i.e., the earlier it should
	// be executed
	priority int
}

func constructAuthenticatorListItem(authenticator Authenticator, priority int) authenticatorListItem {
	return authenticatorListItem{authenticator: authenticator, priority: priority}
}

type authenticatorList []authenticatorListItem

func newAuthenticatorList() *authenticatorList {
	ret := authenticatorList(make([]authenticatorListItem, 0))
	return &ret
}

func (l *authenticatorList) Len() int { return len(*l) }

func (l *authenticatorList) Less(i, j int) bool {
	al := *l
	return al[i].priority < al[j].priority // lower number (higher priority) at front
}

func (l *authenticatorList) Swap(i, j int) {
	al := *l
	al[i], al[j] = al[j], al[i]
}

func (l *authenticatorList) Push(x authenticatorListItem) {
	*l = append(*l, x)
	sort.Sort(l)
}

func (l *authenticatorList) Iterate(agentName string, agentID string, token string, responder *Responder, connCtx ConnContext) HandleCode {
	al := *l
	ret := DECLINED
	for _, item := range al {
		ret = item.authenticator.Authenticate(agentName, agentID, token, responder, connCtx)
		if OK == ret || FAIL == ret {
			break
		}
	}
	return ret
}
