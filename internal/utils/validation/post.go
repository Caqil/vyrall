package validation

import (
	"regexp"
	"strings"
)

// PostContentOptions defines options for post content validation
type PostContentOptions struct {
	MinLength         int
	MaxLength         int
	AllowLinks        bool
	AllowMentions     bool
	AllowHashtags     bool
	MaxMediaItems     int
	MaxMentions       int
	MaxHashtags       int
	DisallowedStrings []string
}

// DefaultPostContentOptions returns default post validation options
func DefaultPostContentOptions() PostContentOptions {
	return PostContentOptions{
		MinLength:         1,
		MaxLength:         5000,
		AllowLinks:        true,
		AllowMentions:     true,
		AllowHashtags:     true,
		MaxMediaItems:     10,
		MaxMentions:       20,
		MaxHashtags:       30,
		DisallowedStrings: []string{},
	}
}

// ValidatePostContent validates post content against the specified options
func ValidatePostContent(content string, options PostContentOptions) (bool, string) {
	// Check length
	contentLen := len(strings.TrimSpace(content))
	if contentLen < options.MinLength {
		return false, "Post content is too short"
	}

	if contentLen > options.MaxLength {
		return false, "Post content exceeds maximum length"
	}

	// Check for disallowed strings
	for _, disallowed := range options.DisallowedStrings {
		if strings.Contains(strings.ToLower(content), strings.ToLower(disallowed)) {
			return false, "Post contains disallowed content"
		}
	}

	// Check for links if not allowed
	if !options.AllowLinks {
		linkPattern := `https?://\S+`
		if MatchesPattern(content, linkPattern) {
			return false, "Links are not allowed in posts"
		}
	}

	// Check for mentions if not allowed or exceeding limit
	if !options.AllowMentions || options.MaxMentions > 0 {
		mentionPattern := `@\w+`
		mentionRegex := regexp.MustCompile(mentionPattern)
		mentions := mentionRegex.FindAllString(content, -1)

		if !options.AllowMentions && len(mentions) > 0 {
			return false, "Mentions are not allowed in posts"
		}

		if options.MaxMentions > 0 && len(mentions) > options.MaxMentions {
			return false, "Post exceeds maximum number of mentions"
		}
	}

	// Check for hashtags if not allowed or exceeding limit
	if !options.AllowHashtags || options.MaxHashtags > 0 {
		hashtagPattern := `#\w+`
		hashtagRegex := regexp.MustCompile(hashtagPattern)
		hashtags := hashtagRegex.FindAllString(content, -1)

		if !options.AllowHashtags && len(hashtags) > 0 {
			return false, "Hashtags are not allowed in posts"
		}

		if options.MaxHashtags > 0 && len(hashtags) > options.MaxHashtags {
			return false, "Post exceeds maximum number of hashtags"
		}
	}

	return true, ""
}

// ExtractHashtags extracts hashtags from post content
func ExtractHashtags(content string) []string {
	hashtagPattern := `#(\w+)`
	hashtagRegex := regexp.MustCompile(hashtagPattern)
	matches := hashtagRegex.FindAllStringSubmatch(content, -1)

	hashtags := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			hashtags = append(hashtags, match[1])
		}
	}

	return hashtags
}

// ExtractMentions extracts user mentions from post content
func ExtractMentions(content string) []string {
	mentionPattern := `@(\w+)`
	mentionRegex := regexp.MustCompile(mentionPattern)
	matches := mentionRegex.FindAllStringSubmatch(content, -1)

	mentions := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			mentions = append(mentions, match[1])
		}
	}

	return mentions
}

// ExtractLinks extracts URLs from post content
func ExtractLinks(content string) []string {
	linkPattern := `https?://\S+`
	linkRegex := regexp.MustCompile(linkPattern)
	links := linkRegex.FindAllString(content, -1)

	// Trim trailing punctuation
	for i, link := range links {
		links[i] = strings.TrimRight(link, ",.!?:;'\"()[]{}")
	}

	return links
}

// ValidatePostTitle validates a post title
func ValidatePostTitle(title string) (bool, string) {
	// Check if title is empty
	if IsEmpty(title) {
		return false, "Title cannot be empty"
	}

	// Check title length
	titleLen := len(strings.TrimSpace(title))
	if titleLen < 1 {
		return false, "Title is too short"
	}

	if titleLen > 300 {
		return false, "Title exceeds maximum length of 300 characters"
	}

	return true, ""
}

// SanitizePostContent sanitizes post content by removing harmful elements
func SanitizePostContent(content string) string {
	// Sanitize content
	sanitized := SanitizeString(content)

	// Additional sanitization specific to posts could be added here

	return sanitized
}

// ValidatePostMedia validates the media items attached to a post
func ValidatePostMedia(mediaCount int, mediaTypes []string) (bool, string) {
	options := DefaultPostContentOptions()

	// Check media count
	if mediaCount > options.MaxMediaItems {
		return false, "Post exceeds maximum number of media items"
	}

	// Check for unsupported media types
	for _, mediaType := range mediaTypes {
		switch mediaType {
		case "image", "video", "audio", "document":
			// Supported types
		default:
			return false, "Post contains unsupported media type: " + mediaType
		}
	}

	return true, ""
}
