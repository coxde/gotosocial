// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gtsmodel

// A policy entry corresponds to
// one ActivityPub URI representing
// an Actor or a Collection of Actors.
type PolicyEntry int

const (
	// IMPORTANT: If adding policy entry values,
	// add them *TO THE BOTTOM OF THE LIST*, and
	// *DO NOT CHANGE THE ORDER OF THE LIST*, as
	// these values are stored in the database,
	// and changing their order changes their
	// meaning and will cause huge problems.

	// ActivityStreams magic public URI, which
	// encompasses every possible Actor URI.
	PolicyEntryPublic PolicyEntry = iota
	// Actor URIs in the Followers Collection
	// of the item owner's Actor.
	PolicyEntryFollowers
	// Actor URIs in the Following Collection
	// of the item owner's Actor.
	PolicyEntryFollowing
	// Actor URIs in the Mutuals Collection
	// of the item owner's Actor.
	//
	// (TODO: Reserved, currently unused).
	PolicyEntryMutuals
	// Actor URIs mentioned/tagged in the item.
	PolicyEntryMentioned
	// Item owner's own Actor URI.
	PolicyEntrySelf
)

// FeasibleForVisibility returns true if the policy entry could
// feasibly be set in a policy for an item with the given visibility,
// otherwise returns false.
//
// For example, PolicyEntryPublic could not be set in a policy for an
// item with visibility FollowersOnly, but could be set in a policy
// for an item with visibility Public or Unlocked.
//
// This is not prescriptive, and should be used only to guide policy
// choices. Eg., if a remote instance wants to do something wacky like
// set "anyone can interact with this status" for a Direct visibility
// status, that's their business; our normal visibility filtering will
// prevent users on our instance from actually being able to interact
// unless they can see the status anyway.
func (p PolicyEntry) FeasibleForVisibility(v Visibility) bool {
	switch p {

	// Mentioned and self are always
	// feasible for any visibility.
	case PolicyEntrySelf,
		PolicyEntryMentioned:
		return true

	// Followers/following/mutual entries
	// are only feasible for items with
	// followers visibility and higher.
	case PolicyEntryFollowers,
		PolicyEntryFollowing:
		return v == VisibilityFollowersOnly ||
			v == VisibilityPublic ||
			v == VisibilityUnlocked

	// Public policy entry only feasible
	// for items that are To or CC public.
	case PolicyEntryPublic:
		return v == VisibilityUnlocked ||
			v == VisibilityPublic

	// Any other combo
	// is probably fine.
	default:
		return true
	}
}

type PolicyEntries []PolicyEntry

// PolicyResult represents the result of
// checking an Actor URI and interaction
// type against the conditions of an
// InteractionPolicy to determine if that
// interaction is permitted.
type PolicyResult int

const (
	// Interaction is not permitted for this
	// Actor URI / interaction combination.
	PolicyResultNo PolicyEntry = iota
	// Interaction is permitted for this Actor
	// URI / interaction combination, but
	// only pending approval by the item owner.
	PolicyResultWithApproval
	// Interaction is permitted for this
	// Actor URI / interaction combination.
	PolicyResultYes
)

// An InteractionPolicy determines which
// interactions will be accepted for an
// item (status), and under what conditions.
type InteractionPolicy struct {
	// Conditions in which a Like
	// interaction will be accepted
	// for an item with this policy.
	CanLike PolicyConditions
	// Conditions in which a Reply
	// interaction will be accepted
	// for an item with this policy.
	CanReply PolicyConditions
	// Conditions in which an Announce
	// interaction will be accepted
	// for an item with this policy.
	CanAnnounce PolicyConditions
}

// PolicyConditions represents the conditions
// in which a certain interaction is permitted
// for various Actors and Actor Collections.
type PolicyConditions struct {
	// Yes is for PolicyEntries of Actor URIs
	// who are permitted to do an interaction
	// without requiring approval.
	Yes PolicyEntries
	// WithApproval is for PolicyEntries of
	// Actor URIs who are permitted to do an
	// interaction only pending approval.
	WithApproval PolicyEntries
}

// Returns the default interaction policy
// for the given visibility level.
func DefaultInteractionPolicyFor(v Visibility) *InteractionPolicy {
	switch v {
	case VisibilityPublic:
		return DefaultInteractionPolicyPublic()
	case VisibilityUnlocked:
		return DefaultInteractionPolicyUnlocked()
	case VisibilityFollowersOnly, VisibilityMutualsOnly:
		return DefaultInteractionPolicyFollowersOnly()
	case VisibilityDirect:
		return DefaultInteractionPolicyDirect()
	default:
		panic("visibility " + v + " not recognized")
	}
}

// Returns the default interaction policy
// for a post with visibility of public.
func DefaultInteractionPolicyPublic() *InteractionPolicy {
	// Anyone can like.
	canLike := make(PolicyEntries, 1)
	canLike[0] = PolicyEntryPublic

	// Unused, set empty.
	canLikeWithApproval := make(PolicyEntries, 0)

	// Anyone can reply.
	canReply := make(PolicyEntries, 1)
	canReply[0] = PolicyEntryPublic

	// Unused, set empty.
	canReplyWithApproval := make(PolicyEntries, 0)

	// Anyone can announce.
	canAnnounce := make(PolicyEntries, 1)
	canAnnounce[0] = PolicyEntryPublic

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyEntries, 0)

	return &InteractionPolicy{
		CanLike: PolicyConditions{
			Yes:          canLike,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyConditions{
			Yes:          canReply,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyConditions{
			Yes:          canAnnounce,
			WithApproval: canAnnounceWithApproval,
		},
	}
}

// Returns the default interaction policy
// for a post with visibility of unlocked.
func DefaultInteractionPolicyUnlocked() *InteractionPolicy {
	// Same as public (for now).
	return DefaultInteractionPolicyPublic()
}

// Returns the default interaction policy for
// a post with visibility of followers only.
func DefaultInteractionPolicyFollowersOnly() *InteractionPolicy {
	// Followers, mentioned, and self can like.
	canLike := make(PolicyEntries, 3)
	canLike[0] = PolicyEntryFollowers
	canLike[1] = PolicyEntryMentioned
	canLike[2] = PolicyEntrySelf

	// Unused, set empty.
	canLikeWithApproval := make(PolicyEntries, 0)

	// Followers, mentioned, and self can reply.
	canReply := make(PolicyEntries, 3)
	canReply[0] = PolicyEntryFollowers
	canReply[1] = PolicyEntryMentioned
	canReply[2] = PolicyEntrySelf

	// Unused, set empty.
	canReplyWithApproval := make(PolicyEntries, 0)

	// Only self can announce.
	canAnnounce := make(PolicyEntries, 1)
	canAnnounce[0] = PolicyEntrySelf

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyEntries, 0)

	return &InteractionPolicy{
		CanLike: PolicyConditions{
			Yes:          canLike,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyConditions{
			Yes:          canReply,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyConditions{
			Yes:          canAnnounce,
			WithApproval: canAnnounceWithApproval,
		},
	}
}

// Returns the default interaction policy
// for a post with visibility of direct.
func DefaultInteractionPolicyDirect() *InteractionPolicy {
	// Mentioned and self can always like.
	canLike := make(PolicyEntries, 2)
	canLike[0] = PolicyEntryMentioned
	canLike[1] = PolicyEntrySelf

	// Unused, set empty.
	canLikeWithApproval := make(PolicyEntries, 0)

	// Mentioned and self can always reply.
	canReply := make(PolicyEntries, 2)
	canReply[0] = PolicyEntryMentioned
	canReply[1] = PolicyEntrySelf

	// Unused, set empty.
	canReplyWithApproval := make(PolicyEntries, 0)

	// Nobody can announce.
	canAnnounce := make(PolicyEntries, 0)
	canAnnounceWithApproval := make(PolicyEntries, 0)

	return &InteractionPolicy{
		CanLike: PolicyConditions{
			Yes:          canLike,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyConditions{
			Yes:          canReply,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyConditions{
			Yes:          canAnnounce,
			WithApproval: canAnnounceWithApproval,
		},
	}
}
