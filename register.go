package tootecho

import (
	"github.com/benpate/toot"
	"github.com/benpate/toot/route"
	"github.com/benpate/toot/scope"
	"github.com/labstack/echo/v4"
)

// Register adds a route to the Echo server for every handler that the API defines.
// A handler left nil is skipped, so an incomplete API registers only what it
// implements. Any middleware passed here runs on each registered route, after the
// built-in WithHost.
//
// Handlers report failure by returning a derp error, which Echo's default error
// handler renders as a 500. Install a derp-aware HTTPErrorHandler on the server to
// get the intended status codes.
func Register[AuthToken toot.ScopesGetter](e *echo.Echo, api toot.API[AuthToken], middleware ...echo.MiddlewareFunc) {

	// https://docs.joinmastodon.org/methods/accounts/
	single_result(api, e.POST, route.PostAccount, api.PostAccount, scope.PostAccount, middleware...)
	single_result(api, e.GET, route.GetAccount_VerifyCredentials, api.GetAccount_VerifyCredentials, scope.GetAccount_VerifyCredentials, middleware...)
	single_result(api, e.PATCH, route.PatchAccount_UpdateCredentials, api.PatchAccount_UpdateCredentials, scope.PatchAccount_UpdateCredentials, middleware...)
	single_result(api, e.GET, route.GetAccounts, api.GetAccounts, scope.GetAccounts, middleware...)
	single_result(api, e.GET, route.GetAccount, api.GetAccount, scope.GetAccount, middleware...)
	paged_result(api, e.GET, route.GetAccount_Statuses, api.GetAccount_Statuses, scope.GetAccount_Statuses, middleware...)
	paged_result(api, e.GET, route.GetAccount_Followers, api.GetAccount_Followers, scope.GetAccount_Followers, middleware...)
	paged_result(api, e.GET, route.GetAccount_Following, api.GetAccount_Following, scope.GetAccount_Following, middleware...)
	paged_result(api, e.GET, route.GetAccount_FeaturedTags, api.GetAccount_FeaturedTags, scope.GetAccount_FeaturedTags, middleware...)
	paged_result(api, e.GET, route.GetAccount_Endorsements, api.GetAccount_Endorsements, scope.GetAccount_Endorsements, middleware...)
	single_result(api, e.GET, route.GetAccount_Lists, api.GetAccount_Lists, scope.GetAccount_Lists, middleware...)
	single_result(api, e.POST, route.PostAccount_Follow, api.PostAccount_Follow, scope.PostAccount_Follow, middleware...)
	single_result(api, e.POST, route.PostAccount_Unfollow, api.PostAccount_Unfollow, scope.PostAccount_Unfollow, middleware...)
	single_result(api, e.POST, route.PostAccount_RemoveFromFollowers, api.PostAccount_RemoveFromFollowers, scope.PostAccount_RemoveFromFollowers, middleware...)
	single_result(api, e.POST, route.PostAccount_Block, api.PostAccount_Block, scope.PostAccount_Block, middleware...)
	single_result(api, e.POST, route.PostAccount_Unblock, api.PostAccount_Unblock, scope.PostAccount_Unblock, middleware...)
	single_result(api, e.POST, route.PostAccount_Mute, api.PostAccount_Mute, scope.PostAccount_Mute, middleware...)
	single_result(api, e.POST, route.PostAccount_Unmute, api.PostAccount_Unmute, scope.PostAccount_Unmute, middleware...)
	single_result(api, e.POST, route.PostAccount_Pin, api.PostAccount_Pin, scope.PostAccount_Pin, middleware...)
	single_result(api, e.POST, route.PostAccount_Unpin, api.PostAccount_Unpin, scope.PostAccount_Unpin, middleware...)
	single_result(api, e.POST, route.PostAccount_Note, api.PostAccount_Note, scope.PostAccount_Note, middleware...)
	single_result(api, e.GET, route.GetAccount_Relationships, api.GetAccount_Relationships, scope.GetAccount_Relationships, middleware...)
	single_result(api, e.GET, route.GetAccount_FamiliarFollowers, api.GetAccount_FamiliarFollowers, scope.GetAccount_FamiliarFollowers, middleware...)
	paged_result(api, e.GET, route.GetAccount_Search, api.GetAccount_Search, scope.GetAccount_Search, middleware...)
	single_result(api, e.GET, route.GetAccount_Lookup, api.GetAccount_Lookup, scope.GetAccount_Lookup, middleware...)

	// https://docs.joinmastodon.org/methods/apps/
	single_result(api, e.POST, route.PostApplication, api.PostApplication, scope.PostApplication, middleware...)
	single_result(api, e.GET, route.GetApplication_VerifyCredentials, api.GetApplication_VerifyCredentials, scope.GetApplication_VerifyCredentials, middleware...)

	// https://docs.joinmastodon.org/methods/announcements/
	single_result(api, e.GET, route.GetAnnouncements, api.GetAnnouncements, scope.GetAnnouncements, middleware...)
	single_result(api, e.POST, route.PostAnnouncement_Dismiss, api.PostAnnouncement_Dismiss, scope.PostAnnouncement_Dismiss, middleware...)
	single_result(api, e.PUT, route.PutAnnouncement_Reaction, api.PutAnnouncement_Reaction, scope.PutAnnouncement_Reaction, middleware...)
	single_result(api, e.DELETE, route.DeleteAnnouncement_Reaction, api.DeleteAnnouncement_Reaction, scope.DeleteAnnouncement_Reaction, middleware...)

	// https://docs.joinmastodon.org/methods/blocks/
	paged_result(api, e.GET, route.GetBlocks, api.GetBlocks, scope.GetBlocks, middleware...)

	// https://docs.joinmastodon.org/methods/bookmarks/
	paged_result(api, e.GET, route.GetBookmarks, api.GetBookmarks, scope.GetBookmarks, middleware...)

	// https://docs.joinmastodon.org/methods/conversations/
	paged_result(api, e.GET, route.GetConversations, api.GetConversations, scope.GetConversations, middleware...)
	single_result(api, e.DELETE, route.DeleteConversation, api.DeleteConversation, scope.DeleteConversation, middleware...)
	single_result(api, e.POST, route.PostConversationRead, api.PostConversationRead, scope.PostConversationRead, middleware...)

	// https://docs.joinmastodon.org/methods/custom_emojis/
	single_result(api, e.GET, route.GetCustomEmojis, api.GetCustomEmojis, scope.GetCustomEmojis, middleware...)

	// https://docs.joinmastodon.org/methods/directory/
	paged_result(api, e.GET, route.GetDirectory, api.GetDirectory, scope.GetDirectory, middleware...)

	// https://docs.joinmastodon.org/methods/domain_blocks/
	paged_result(api, e.GET, route.GetDomainBlocks, api.GetDomainBlocks, scope.GetDomainBlocks, middleware...)
	single_result(api, e.POST, route.PostDomainBlock, api.PostDomainBlock, scope.PostDomainBlock, middleware...)
	single_result(api, e.DELETE, route.DeleteDomainBlock, api.DeleteDomainBlock, scope.DeleteDomainBlock, middleware...)

	// https://docs.joinmastodon.org/methods/emails/
	single_result(api, e.POST, route.PostEmailConfirmation, api.PostEmailConfirmation, scope.PostEmailConfirmation, middleware...)

	// https://docs.joinmastodon.org/methods/endorsements/
	paged_result(api, e.GET, route.GetEndorsements, api.GetEndorsements, scope.GetEndorsements, middleware...)

	// https://docs.joinmastodon.org/methods/favourites/
	paged_result(api, e.GET, route.GetFavourites, api.GetFavourites, scope.GetFavourites, middleware...)

	// https://docs.joinmastodon.org/methods/featured_tags/
	single_result(api, e.GET, route.GetFeaturedTags, api.GetFeaturedTags, scope.GetFeaturedTags, middleware...)
	single_result(api, e.POST, route.PostFeaturedTag, api.PostFeaturedTag, scope.PostFeaturedTag, middleware...)
	single_result(api, e.DELETE, route.DeleteFeaturedTag, api.DeleteFeaturedTag, scope.DeleteFeaturedTag, middleware...)
	single_result(api, e.GET, route.GetFeaturedTags_Suggestions, api.GetFeaturedTags_Suggestions, scope.GetFeaturedTags_Suggestions, middleware...)

	// https://docs.joinmastodon.org/methods/filters/
	single_result(api, e.GET, route.GetFilters, api.GetFilters, scope.GetFilters, middleware...)
	single_result(api, e.GET, route.GetFilter, api.GetFilter, scope.GetFilter, middleware...)
	single_result(api, e.POST, route.PostFilter, api.PostFilter, scope.PostFilter, middleware...)
	single_result(api, e.PUT, route.PutFilter, api.PutFilter, scope.PutFilter, middleware...)
	single_result(api, e.DELETE, route.DeleteFilter, api.DeleteFilter, scope.DeleteFilter, middleware...)
	single_result(api, e.GET, route.GetFilter_Keywords, api.GetFilter_Keywords, scope.GetFilter_Keywords, middleware...)
	single_result(api, e.POST, route.PostFilter_Keyword, api.PostFilter_Keyword, scope.PostFilter_Keyword, middleware...)
	single_result(api, e.GET, route.GetFilter_Keyword, api.GetFilter_Keyword, scope.GetFilter_Keyword, middleware...)
	single_result(api, e.PUT, route.PutFilter_Keyword, api.PutFilter_Keyword, scope.PutFilter_Keyword, middleware...)
	single_result(api, e.DELETE, route.DeleteFilter_Keyword, api.DeleteFilter_Keyword, scope.DeleteFilter_Keyword, middleware...)
	single_result(api, e.GET, route.GetFilter_Statuses, api.GetFilter_Statuses, scope.GetFilter_Statuses, middleware...)
	single_result(api, e.POST, route.PostFilter_Status, api.PostFilter_Status, scope.PostFilter_Status, middleware...)
	single_result(api, e.GET, route.GetFilter_Status, api.GetFilter_Status, scope.GetFilter_Status, middleware...)
	single_result(api, e.DELETE, route.DeleteFilter_Status, api.DeleteFilter_Status, scope.DeleteFilter_Status, middleware...)
	paged_result(api, e.GET, route.GetFilters_V1, api.GetFilters_V1, scope.GetFilters_V1, middleware...)
	single_result(api, e.GET, route.GetFilter_V1, api.GetFilter_V1, scope.GetFilter_V1, middleware...)
	single_result(api, e.POST, route.PostFilter_V1, api.PostFilter_V1, scope.PostFilter_V1, middleware...)
	single_result(api, e.PUT, route.PutFilter_V1, api.PutFilter_V1, scope.PutFilter_V1, middleware...)
	single_result(api, e.DELETE, route.DeleteFilter_V1, api.DeleteFilter_V1, scope.DeleteFilter_V1, middleware...)

	// https://docs.joinmastodon.org/methods/follow_requests/
	paged_result(api, e.GET, route.GetFollowRequests, api.GetFollowRequests, scope.GetFollowRequests, middleware...)
	single_result(api, e.POST, route.PostFollowRequest_Authorize, api.PostFollowRequest_Authorize, scope.PostFollowRequest_Authorize, middleware...)
	single_result(api, e.POST, route.PostFollowRequest_Reject, api.PostFollowRequest_Reject, scope.PostFollowRequest_Reject, middleware...)

	// https://docs.joinmastodon.org/methods/followed_tags/
	paged_result(api, e.GET, route.GetFollowedTags, api.GetFollowedTags, scope.GetFollowedTags, middleware...)

	// https://docs.joinmastodon.org/methods/instance/
	single_result(api, e.GET, route.GetInstance, api.GetInstance, scope.GetInstance, middleware...)
	single_result(api, e.GET, route.GetInstance_Peers, api.GetInstance_Peers, scope.GetInstance_Peers, middleware...)
	single_result(api, e.GET, route.GetInstance_Activity, api.GetInstance_Activity, scope.GetInstance_Activity, middleware...)
	single_result(api, e.GET, route.GetInstance_Rules, api.GetInstance_Rules, scope.GetInstance_Rules, middleware...)
	single_result(api, e.GET, route.GetInstance_DomainBlocks, api.GetInstance_DomainBlocks, scope.GetInstance_DomainBlocks, middleware...)
	single_result(api, e.GET, route.GetInstance_ExtendedDescription, api.GetInstance_ExtendedDescription, scope.GetInstance_ExtendedDescription, middleware...)
	single_result(api, e.GET, route.GetInstance_V1, api.GetInstance_V1, scope.GetInstance_V1, middleware...)

	// https://docs.joinmastodon.org/methods/lists/
	single_result(api, e.GET, route.GetLists, api.GetLists, scope.GetLists, middleware...)
	single_result(api, e.GET, route.GetList, api.GetList, scope.GetList, middleware...)
	single_result(api, e.POST, route.PostList, api.PostList, scope.PostList, middleware...)
	single_result(api, e.PUT, route.PutList, api.PutList, scope.PutList, middleware...)
	single_result(api, e.DELETE, route.DeleteList, api.DeleteList, scope.DeleteList, middleware...)
	paged_result(api, e.GET, route.GetList_Accounts, api.GetList_Accounts, scope.GetList_Accounts, middleware...)
	single_result(api, e.POST, route.PostList_Accounts, api.PostList_Accounts, scope.PostList_Accounts, middleware...)
	single_result(api, e.DELETE, route.DeleteList_Accounts, api.DeleteList_Accounts, scope.DeleteList_Accounts, middleware...)

	// https://docs.joinmastodon.org/methods/markers/
	single_result(api, e.GET, route.GetMarkers, api.GetMarkers, scope.GetMarkers, middleware...)
	single_result(api, e.POST, route.PostMarker, api.PostMarker, scope.PostMarker, middleware...)

	// https://docs.joinmastodon.org/methods/media/
	single_result(api, e.POST, route.PostMedia, api.PostMedia, scope.PostMedia, middleware...)
	single_result(api, e.GET, route.GetMedia, api.GetMedia, scope.GetMedia, middleware...)
	single_result(api, e.PUT, route.PutMedia, api.PutMedia, scope.PutMedia, middleware...)

	// https://docs.joinmastodon.org/methods/mutes/
	paged_result(api, e.GET, route.GetMutes, api.GetMutes, scope.GetMutes, middleware...)

	// https://docs.joinmastodon.org/methods/notifications/
	paged_result(api, e.GET, route.GetNotifications, api.GetNotifications, scope.GetNotifications, middleware...)
	single_result(api, e.GET, route.GetNotification, api.GetNotification, scope.GetNotification, middleware...)
	single_result(api, e.POST, route.PostNotifications_Clear, api.PostNotifications_Clear, scope.PostNotifications_Clear, middleware...)
	single_result(api, e.POST, route.PostNotification_Dismiss, api.PostNotification_Dismiss, scope.PostNotification_Dismiss, middleware...)
	single_result(api, e.GET, route.GetNotifications_UnreadCount, api.GetNotifications_UnreadCount, scope.GetNotifications_UnreadCount, middleware...)
	single_result(api, e.GET, route.GetNotificationPolicy, api.GetNotificationPolicy, scope.GetNotificationPolicy, middleware...)

	// https://docs.joinmastodon.org/methods/oauth/
	single_result(api, e.GET, route.GetOAuth_Authorize, api.GetOAuth_Authorize, scope.GetOAuth_Authorize, middleware...)
	single_result(api, e.POST, route.PostOAuth_Token, api.PostOAuth_Token, scope.PostOAuth_Token, middleware...)
	single_result(api, e.POST, route.PostOAuth_Revoke, api.PostOAuth_Revoke, scope.PostOAuth_Revoke, middleware...)

	// https://docs.joinmastodon.org/methods/oembed/
	single_result(api, e.GET, route.GetOEmbed, api.GetOEmbed, scope.GetOEmbed, middleware...)

	// https://docs.joinmastodon.org/methods/polls/
	single_result(api, e.GET, route.GetPoll, api.GetPoll, scope.GetPoll, middleware...)
	single_result(api, e.POST, route.PostPoll_Votes, api.PostPoll_Votes, scope.PostPoll_Votes, middleware...)

	// https://docs.joinmastodon.org/methods/preferences/
	single_result(api, e.GET, route.GetPreferences, api.GetPreferences, scope.GetPreferences, middleware...)

	// https://docs.joinmastodon.org/methods/push/
	single_result(api, e.POST, route.PostPushSubscription, api.PostPushSubscription, scope.PostPushSubscription, middleware...)
	single_result(api, e.GET, route.GetPushSubscription, api.GetPushSubscription, scope.GetPushSubscription, middleware...)
	single_result(api, e.PUT, route.PutPushSubscription, api.PutPushSubscription, scope.PutPushSubscription, middleware...)
	single_result(api, e.DELETE, route.DeletePushSubscription, api.DeletePushSubscription, scope.DeletePushSubscription, middleware...)

	// https://docs.joinmastodon.org/methods/profile/
	single_result(api, e.DELETE, route.DeleteProfile_Avatar, api.DeleteProfile_Avatar, scope.DeleteProfile_Avatar, middleware...)
	single_result(api, e.DELETE, route.DeleteProfile_Header, api.DeleteProfile_Header, scope.DeleteProfile_Header, middleware...)

	// https://docs.joinmastodon.org/methods/reports/
	single_result(api, e.POST, route.PostReport, api.PostReport, scope.PostReport, middleware...)

	// https://docs.joinmastodon.org/methods/scheduled_statuses/
	paged_result(api, e.GET, route.GetScheduledStatuses, api.GetScheduledStatuses, scope.GetScheduledStatuses, middleware...)
	single_result(api, e.GET, route.GetScheduledStatus, api.GetScheduledStatus, scope.GetScheduledStatus, middleware...)
	single_result(api, e.PUT, route.PutScheduledStatus, api.PutScheduledStatus, scope.PutScheduledStatus, middleware...)
	single_result(api, e.DELETE, route.DeleteScheduledStatus, api.DeleteScheduledStatus, scope.DeleteScheduledStatus, middleware...)

	// https://docs.joinmastodon.org/methods/search/
	single_result(api, e.GET, route.GetSearch, api.GetSearch, scope.GetSearch, middleware...)

	// https://docs.joinmastodon.org/methods/statuses/
	single_result(api, e.POST, route.PostStatus, api.PostStatus, scope.PostStatus, middleware...)
	single_result(api, e.GET, route.GetStatus, api.GetStatus, scope.GetStatus, middleware...)
	single_result(api, e.DELETE, route.DeleteStatus, api.DeleteStatus, scope.DeleteStatus, middleware...)
	single_result(api, e.GET, route.GetStatus_Context, api.GetStatus_Context, scope.GetStatus_Context, middleware...)
	single_result(api, e.POST, route.PostStatus_Translate, api.PostStatus_Translate, scope.PostStatus_Translate, middleware...)
	paged_result(api, e.GET, route.GetStatus_RebloggedBy, api.GetStatus_RebloggedBy, scope.GetStatus_RebloggedBy, middleware...)
	paged_result(api, e.GET, route.GetStatus_FavouritedBy, api.GetStatus_FavouritedBy, scope.GetStatus_FavouritedBy, middleware...)
	single_result(api, e.POST, route.PostStatus_Favourite, api.PostStatus_Favourite, scope.PostStatus_Favourite, middleware...)
	single_result(api, e.POST, route.PostStatus_Unfavourite, api.PostStatus_Unfavourite, scope.PostStatus_Unfavourite, middleware...)
	single_result(api, e.POST, route.PostStatus_Reblog, api.PostStatus_Reblog, scope.PostStatus_Reblog, middleware...)
	single_result(api, e.POST, route.PostStatus_Unreblog, api.PostStatus_Unreblog, scope.PostStatus_Unreblog, middleware...)
	single_result(api, e.POST, route.PostStatus_Bookmark, api.PostStatus_Bookmark, scope.PostStatus_Bookmark, middleware...)
	single_result(api, e.POST, route.PostStatus_Unbookmark, api.PostStatus_Unbookmark, scope.PostStatus_Unbookmark, middleware...)
	single_result(api, e.POST, route.PostStatus_Mute, api.PostStatus_Mute, scope.PostStatus_Mute, middleware...)
	single_result(api, e.POST, route.PostStatus_Unmute, api.PostStatus_Unmute, scope.PostStatus_Unmute, middleware...)
	single_result(api, e.POST, route.PostStatus_Pin, api.PostStatus_Pin, scope.PostStatus_Pin, middleware...)
	single_result(api, e.POST, route.PostStatus_Unpin, api.PostStatus_Unpin, scope.PostStatus_Unpin, middleware...)
	single_result(api, e.PUT, route.PutStatus, api.PutStatus, scope.PutStatus, middleware...)
	single_result(api, e.GET, route.GetStatus_History, api.GetStatus_History, scope.GetStatus_History, middleware...)
	single_result(api, e.GET, route.GetStatus_Source, api.GetStatus_Source, scope.GetStatus_Source, middleware...)

	// https://docs.joinmastodon.org/methods/suggestions/
	single_result(api, e.GET, route.GetSuggestions, api.GetSuggestions, scope.GetSuggestions, middleware...)
	single_result(api, e.DELETE, route.DeleteSuggestion, api.DeleteSuggestion, scope.DeleteSuggestion, middleware...)

	// https://docs.joinmastodon.org/methods/tags/
	single_result(api, e.GET, route.GetTag, api.GetTag, scope.GetTag, middleware...)
	single_result(api, e.POST, route.PostTag_Follow, api.PostTag_Follow, scope.PostTag_Follow, middleware...)
	single_result(api, e.POST, route.PostTag_Unfollow, api.PostTag_Unfollow, scope.PostTag_Unfollow, middleware...)

	// https://docs.joinmastodon.org/methods/timelines/
	paged_result(api, e.GET, route.GetTimeline_Public, api.GetTimeline_Public, scope.GetTimeline_Public, middleware...)
	paged_result(api, e.GET, route.GetTimeline_Hashtag, api.GetTimeline_Hashtag, scope.GetTimeline_Hashtag, middleware...)
	paged_result(api, e.GET, route.GetTimeline_Home, api.GetTimeline_Home, scope.GetTimeline_Home, middleware...)
	paged_result(api, e.GET, route.GetTimeline_List, api.GetTimeline_List, scope.GetTimeline_List, middleware...)

	// https://docs.joinmastodon.org/methods/trends/
	paged_result(api, e.GET, route.GetTrends, api.GetTrends, scope.GetTrends, middleware...)
	paged_result(api, e.GET, route.GetTrends_Statuses, api.GetTrends_Statuses, scope.GetTrends_Statuses, middleware...)
	paged_result(api, e.GET, route.GetTrends_Links, api.GetTrends_Links, scope.GetTrends_Links, middleware...)
}
