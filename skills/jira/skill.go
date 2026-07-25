// Package jira provides an omniskill Skill for Jira issue tracking.
//
// This package can be used standalone with mcp-atlassian or composed
// with other skills in a multi-service MCP server.
package jira

import (
	"context"
	"fmt"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/core"
	"github.com/grokify/go-atlassian/jira"
	"github.com/plexusone/omniskill/skill"
)

// Skill provides Jira issue tracking tools.
type Skill struct {
	client *jira.Client
}

// New creates a new Jira skill with the given client.
func New(client *jira.Client) *Skill {
	return &Skill{client: client}
}

func (s *Skill) Name() string        { return "jira" }
func (s *Skill) Description() string { return "Jira issue tracking, agile boards, and reporting" }
func (s *Skill) Init(context.Context) error { return nil }
func (s *Skill) Close() error               { return nil }

var _ skill.Skill = (*Skill)(nil)

func (s *Skill) Tools() []skill.Tool {
	return []skill.Tool{
		s.getIssueTool(),
		s.searchTool(),
		s.updateIssueTool(),
		s.addCommentTool(),
		s.getTransitionsTool(),
		s.transitionIssueTool(),
		s.getCommentsTool(),
		s.getProjectsTool(),
		s.createIssueTool(),
		s.cloneIssueTool(),
		s.bulkUpdateTool(),
		s.getBoardsTool(),
		s.getSprintsTool(),
		s.velocityReportTool(),
		s.burndownReportTool(),
		s.worklogReportTool(),
		s.cycleTimeReportTool(),
		s.moveToSprintTool(),
	}
}

func (s *Skill) getIssueTool() skill.Tool {
	return skill.NewTool(
		"jira_get_issue",
		"Get a Jira issue by key with all fields including description, status, assignee, and custom fields",
		map[string]skill.Parameter{
			"key":    {Type: "string", Description: "Issue key (e.g., PROJ-123)", Required: true},
			"expand": {Type: "string", Description: "Comma-separated list of fields to expand (e.g., changelog,renderedFields)"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			key, _ := params["key"].(string)
			if key == "" {
				return nil, fmt.Errorf("key is required")
			}
			var opts *jira.GetQueryOptions
			if expand, ok := params["expand"].(string); ok && expand != "" {
				opts = &jira.GetQueryOptions{ExpandChangelog: expand == "changelog" || expand == "all"}
			}
			issue, err := s.client.IssueAPI.Issue(ctx, key, opts)
			if err != nil {
				return nil, fmt.Errorf("get issue %s: %w", key, err)
			}
			return jira.ToIssueOutput(issue), nil
		},
	)
}

func (s *Skill) searchTool() skill.Tool {
	return skill.NewTool(
		"jira_search",
		"Search Jira issues using JQL (Jira Query Language). Returns matching issues with key fields.",
		map[string]skill.Parameter{
			"jql":         {Type: "string", Description: "JQL query string (e.g., 'project = PROJ AND status = Open')", Required: true},
			"max_results": {Type: "integer", Description: "Maximum number of results to return (default: 50, max: 100)", Default: 50},
			"fields":      {Type: "string", Description: "Comma-separated list of fields to return"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			jql, _ := params["jql"].(string)
			if jql == "" {
				return nil, fmt.Errorf("jql is required")
			}
			maxResults := 50
			if mr, ok := params["max_results"].(float64); ok {
				maxResults = int(mr)
				if maxResults > 100 {
					maxResults = 100
				}
			}
			issues, err := s.client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
			if err != nil {
				return nil, fmt.Errorf("search failed: %w", err)
			}
			if len(issues) > maxResults {
				issues = issues[:maxResults]
			}
			return map[string]any{"total": len(issues), "issues": jira.ToIssueOutputs(issues)}, nil
		},
	)
}

func (s *Skill) updateIssueTool() skill.Tool {
	return skill.NewTool(
		"jira_update_issue",
		"Update a Jira issue's fields such as summary, description, labels, or custom fields",
		map[string]skill.Parameter{
			"key":           {Type: "string", Description: "Issue key (e.g., PROJ-123)", Required: true},
			"summary":       {Type: "string", Description: "New summary/title for the issue"},
			"description":   {Type: "string", Description: "New description for the issue"},
			"labels":        {Type: "array", Description: "Labels to set on the issue (replaces existing labels)"},
			"add_labels":    {Type: "array", Description: "Labels to add to the issue (preserves existing labels)"},
			"remove_labels": {Type: "array", Description: "Labels to remove from the issue"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			key, _ := params["key"].(string)
			if key == "" {
				return nil, fmt.Errorf("key is required")
			}
			updateBody := jira.IssuePatchRequestBody{}
			hasUpdate := false

			if addLabels, ok := params["add_labels"].([]any); ok && len(addLabels) > 0 {
				if updateBody.Update == nil {
					updateBody.Update = &jira.IssuePatchRequestBodyUpdate{}
				}
				for _, label := range addLabels {
					if labelStr, ok := label.(string); ok {
						labelCopy := labelStr
						updateBody.Update.Labels = append(updateBody.Update.Labels, jira.IssuePatchRequestBodyUpdateLabel{Add: &labelCopy})
					}
				}
				hasUpdate = true
			}
			if removeLabels, ok := params["remove_labels"].([]any); ok && len(removeLabels) > 0 {
				if updateBody.Update == nil {
					updateBody.Update = &jira.IssuePatchRequestBodyUpdate{}
				}
				for _, label := range removeLabels {
					if labelStr, ok := label.(string); ok {
						labelCopy := labelStr
						updateBody.Update.Labels = append(updateBody.Update.Labels, jira.IssuePatchRequestBodyUpdateLabel{Remove: &labelCopy})
					}
				}
				hasUpdate = true
			}
			if summary, ok := params["summary"].(string); ok && summary != "" {
				if updateBody.Fields == nil {
					updateBody.Fields = make(map[string]jira.IssuePatchRequestBodyField)
				}
				updateBody.Fields["summary"] = jira.IssuePatchRequestBodyField{Value: summary}
				hasUpdate = true
			}
			if description, ok := params["description"].(string); ok && description != "" {
				if updateBody.Fields == nil {
					updateBody.Fields = make(map[string]jira.IssuePatchRequestBodyField)
				}
				updateBody.Fields["description"] = jira.IssuePatchRequestBodyField{Value: description}
				hasUpdate = true
			}
			if !hasUpdate {
				return nil, fmt.Errorf("no update fields provided")
			}
			if _, err := s.client.IssueAPI.IssuePatch(ctx, key, updateBody); err != nil {
				return nil, fmt.Errorf("update issue %s: %w", key, err)
			}
			return map[string]any{"success": true, "key": key, "message": "Issue updated successfully"}, nil
		},
	)
}

func (s *Skill) addCommentTool() skill.Tool {
	return skill.NewTool(
		"jira_add_comment",
		"Add a comment to a Jira issue",
		map[string]skill.Parameter{
			"key":  {Type: "string", Description: "Issue key (e.g., PROJ-123)", Required: true},
			"body": {Type: "string", Description: "Comment body text", Required: true},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			key, _ := params["key"].(string)
			body, _ := params["body"].(string)
			if key == "" {
				return nil, fmt.Errorf("key is required")
			}
			if body == "" {
				return nil, fmt.Errorf("body is required")
			}
			comment, _, err := s.client.JiraClient.Issue.AddCommentWithContext(ctx, key, &gojira.Comment{Body: body})
			if err != nil {
				return nil, fmt.Errorf("add comment to %s: %w", key, err)
			}
			return map[string]any{"success": true, "key": key, "comment_id": comment.ID, "message": "Comment added successfully"}, nil
		},
	)
}

func (s *Skill) getTransitionsTool() skill.Tool {
	return skill.NewTool(
		"jira_get_transitions",
		"Get available status transitions for a Jira issue",
		map[string]skill.Parameter{
			"key": {Type: "string", Description: "Issue key (e.g., PROJ-123)", Required: true},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			key, _ := params["key"].(string)
			if key == "" {
				return nil, fmt.Errorf("key is required")
			}
			transitions, _, err := s.client.IssueAPI.GetTransitions(ctx, key, false)
			if err != nil {
				return nil, fmt.Errorf("get transitions for %s: %w", key, err)
			}
			results := make([]map[string]any, 0, len(transitions))
			for _, t := range transitions {
				results = append(results, map[string]any{"id": t.ID, "name": t.Name, "to": t.To.Name})
			}
			return map[string]any{"key": key, "transitions": results}, nil
		},
	)
}

func (s *Skill) transitionIssueTool() skill.Tool {
	return skill.NewTool(
		"jira_transition_issue",
		"Transition a Jira issue to a new status. Use either transition name or ID.",
		map[string]skill.Parameter{
			"key":        {Type: "string", Description: "Issue key (e.g., PROJ-123)", Required: true},
			"transition": {Type: "string", Description: "Transition name (e.g., 'In Progress', 'Done') or ID", Required: true},
			"comment":    {Type: "string", Description: "Optional comment to add with the transition"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			key, _ := params["key"].(string)
			transition, _ := params["transition"].(string)
			if key == "" {
				return nil, fmt.Errorf("key is required")
			}
			if transition == "" {
				return nil, fmt.Errorf("transition is required")
			}
			var opts *jira.TransitionOptions
			if comment, ok := params["comment"].(string); ok && comment != "" {
				opts = &jira.TransitionOptions{Comment: comment}
			}
			result, err := s.client.IssueAPI.TransitionIssue(ctx, key, transition, opts)
			if err != nil {
				return nil, fmt.Errorf("transition issue %s: %w", key, err)
			}
			return map[string]any{
				"success": true, "key": result.Key,
				"from_status": result.FromStatus, "to_status": result.ToStatus,
				"transition": result.Transition,
				"message":    fmt.Sprintf("Issue transitioned from %s to %s", result.FromStatus, result.ToStatus),
			}, nil
		},
	)
}

func (s *Skill) getCommentsTool() skill.Tool {
	return skill.NewTool(
		"jira_get_comments",
		"Get comments on a Jira issue",
		map[string]skill.Parameter{
			"key":         {Type: "string", Description: "Issue key (e.g., PROJ-123)", Required: true},
			"max_results": {Type: "integer", Description: "Maximum number of comments to return (default: 50)", Default: 50},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			key, _ := params["key"].(string)
			if key == "" {
				return nil, fmt.Errorf("key is required")
			}
			maxResults := 50
			if mr, ok := params["max_results"].(float64); ok {
				maxResults = int(mr)
			}
			return s.client.GetComments(ctx, key, maxResults)
		},
	)
}

func (s *Skill) getProjectsTool() skill.Tool {
	return skill.NewTool(
		"jira_get_projects",
		"List available Jira projects",
		map[string]skill.Parameter{},
		func(ctx context.Context, _ map[string]any) (any, error) {
			projects, _, err := s.client.JiraClient.Project.GetListWithContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("get projects: %w", err)
			}
			results := make([]map[string]any, 0, len(*projects))
			for _, p := range *projects {
				results = append(results, map[string]any{"key": p.Key, "name": p.Name, "id": p.ID})
			}
			return map[string]any{"total": len(results), "projects": results}, nil
		},
	)
}

func (s *Skill) createIssueTool() skill.Tool {
	return skill.NewTool(
		"jira_create_issue",
		"Create a new Jira issue (Story, Bug, Task, etc.) with support for custom fields",
		map[string]skill.Parameter{
			"project":       {Type: "string", Description: "Project key (e.g., PROJ)", Required: true},
			"type":          {Type: "string", Description: "Issue type (e.g., Story, Bug, Task, Epic)", Required: true},
			"summary":       {Type: "string", Description: "Issue summary/title", Required: true},
			"description":   {Type: "string", Description: "Issue description (supports Jira markdown)"},
			"parent":        {Type: "string", Description: "Parent issue key for subtasks or stories under epics"},
			"labels":        {Type: "array", Description: "Labels to apply to the issue"},
			"priority":      {Type: "string", Description: "Priority name (e.g., High, Medium, Low)"},
			"assignee":      {Type: "string", Description: "Assignee username or email"},
			"components":    {Type: "array", Description: "Component names"},
			"custom_fields": {Type: "object", Description: "Custom fields as key-value pairs"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			input := &core.IssueInput{}
			input.Project, _ = params["project"].(string)
			input.Type, _ = params["type"].(string)
			input.Summary, _ = params["summary"].(string)
			if input.Project == "" || input.Type == "" || input.Summary == "" {
				return nil, fmt.Errorf("project, type, and summary are required")
			}
			input.Description, _ = params["description"].(string)
			input.Parent, _ = params["parent"].(string)
			input.Priority, _ = params["priority"].(string)
			input.Assignee, _ = params["assignee"].(string)
			if labels, ok := params["labels"].([]any); ok {
				for _, l := range labels {
					if ls, ok := l.(string); ok {
						input.Labels = append(input.Labels, ls)
					}
				}
			}
			if components, ok := params["components"].([]any); ok {
				for _, c := range components {
					if cs, ok := c.(string); ok {
						input.Components = append(input.Components, cs)
					}
				}
			}
			if cf, ok := params["custom_fields"].(map[string]any); ok {
				input.CustomFields = cf
			}
			result, err := core.CreateIssue(ctx, s.client, input)
			if err != nil {
				return nil, fmt.Errorf("create issue: %w", err)
			}
			return map[string]any{
				"success": true, "key": result.Key, "id": result.ID,
				"self": result.Self, "summary": result.Summary,
				"message": fmt.Sprintf("Issue %s created successfully", result.Key),
			}, nil
		},
	)
}

func (s *Skill) cloneIssueTool() skill.Tool {
	return skill.NewTool(
		"jira_clone_issue",
		"Clone a Jira issue with configurable field mapping",
		map[string]skill.Parameter{
			"source_key":       {Type: "string", Description: "Source issue key to clone (e.g., PROJ-123)", Required: true},
			"target_project":   {Type: "string", Description: "Target project key (default: same as source)"},
			"target_type":      {Type: "string", Description: "Target issue type (default: same as source)"},
			"summary_prefix":   {Type: "string", Description: "Prefix to add to cloned issue summary"},
			"summary_suffix":   {Type: "string", Description: "Suffix to add to cloned issue summary"},
			"link_to_original": {Type: "boolean", Description: "Create a link to the original issue (default: false)"},
			"link_type":        {Type: "string", Description: "Link type name when linking to original (default: 'Cloners')"},
			"exclude_fields":   {Type: "array", Description: "Fields to exclude from cloning"},
			"parent":           {Type: "string", Description: "Parent issue key for subtasks"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			sourceKey, _ := params["source_key"].(string)
			if sourceKey == "" {
				return nil, fmt.Errorf("source_key is required")
			}
			opts := &jira.CloneOptions{}
			opts.TargetProject, _ = params["target_project"].(string)
			opts.TargetType, _ = params["target_type"].(string)
			opts.SummaryPrefix, _ = params["summary_prefix"].(string)
			opts.SummarySuffix, _ = params["summary_suffix"].(string)
			opts.LinkToOriginal, _ = params["link_to_original"].(bool)
			opts.LinkType, _ = params["link_type"].(string)
			opts.Parent, _ = params["parent"].(string)
			if fields, ok := params["exclude_fields"].([]any); ok {
				for _, f := range fields {
					if fs, ok := f.(string); ok {
						opts.ExcludeFields = append(opts.ExcludeFields, fs)
					}
				}
			}
			result, err := s.client.IssueAPI.CloneIssue(ctx, sourceKey, opts)
			if err != nil {
				return nil, fmt.Errorf("clone issue %s: %w", sourceKey, err)
			}
			return map[string]any{
				"success": true, "source_key": result.SourceKey, "cloned_key": result.ClonedKey,
				"linked": result.Linked, "message": fmt.Sprintf("Issue %s cloned to %s", result.SourceKey, result.ClonedKey),
			}, nil
		},
	)
}

func (s *Skill) bulkUpdateTool() skill.Tool {
	return skill.NewTool(
		"jira_bulk_update",
		"Update multiple Jira issues with the same field changes",
		map[string]skill.Parameter{
			"issue_keys":    {Type: "array", Description: "Issue keys to update"},
			"jql":           {Type: "string", Description: "JQL query to select issues (alternative to issue_keys)"},
			"add_labels":    {Type: "array", Description: "Labels to add to all issues"},
			"remove_labels": {Type: "array", Description: "Labels to remove from all issues"},
			"comment":       {Type: "string", Description: "Comment to add to all issues"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			var issueKeys []string
			if keys, ok := params["issue_keys"].([]any); ok {
				for _, k := range keys {
					if ks, ok := k.(string); ok {
						issueKeys = append(issueKeys, ks)
					}
				}
			}
			if jql, ok := params["jql"].(string); ok && jql != "" {
				issues, err := s.client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
				if err != nil {
					return nil, fmt.Errorf("JQL search failed: %w", err)
				}
				for _, iss := range issues {
					issueKeys = append(issueKeys, iss.Key)
				}
			}
			if len(issueKeys) == 0 {
				return nil, fmt.Errorf("no issues to update (provide issue_keys or jql)")
			}
			opts := &jira.BulkUpdateOptions{}
			if al, ok := params["add_labels"].([]any); ok {
				for _, l := range al {
					if ls, ok := l.(string); ok {
						opts.AddLabels = append(opts.AddLabels, ls)
					}
				}
			}
			if rl, ok := params["remove_labels"].([]any); ok {
				for _, l := range rl {
					if ls, ok := l.(string); ok {
						opts.RemoveLabels = append(opts.RemoveLabels, ls)
					}
				}
			}
			opts.Comment, _ = params["comment"].(string)
			if len(opts.AddLabels) == 0 && len(opts.RemoveLabels) == 0 && opts.Comment == "" {
				return nil, fmt.Errorf("no updates specified (use add_labels, remove_labels, or comment)")
			}
			result, err := s.client.IssueAPI.BulkUpdateIssues(ctx, issueKeys, opts)
			if err != nil {
				return nil, fmt.Errorf("bulk update failed: %w", err)
			}
			return map[string]any{
				"success": true, "total": result.Total,
				"success_count": result.SuccessCount, "fail_count": result.FailCount,
				"successful": result.Successful, "failed": result.Failed,
				"message": fmt.Sprintf("Updated %d of %d issues", result.SuccessCount, result.Total),
			}, nil
		},
	)
}

func (s *Skill) getBoardsTool() skill.Tool {
	return skill.NewTool(
		"jira_get_boards",
		"List Jira agile boards with optional filtering",
		map[string]skill.Parameter{
			"project": {Type: "string", Description: "Filter boards by project key"},
			"type":    {Type: "string", Description: "Filter boards by type (scrum, kanban)"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			opts := &gojira.BoardListOptions{}
			if project, ok := params["project"].(string); ok {
				opts.ProjectKeyOrID = project
			}
			if boardType, ok := params["type"].(string); ok {
				opts.BoardType = boardType
			}
			boards, err := s.client.BoardAPI.GetBoards(ctx, opts)
			if err != nil {
				return nil, fmt.Errorf("get boards: %w", err)
			}
			results := make([]map[string]any, 0, len(boards))
			for _, b := range boards {
				results = append(results, map[string]any{"id": b.ID, "name": b.Name, "type": b.Type})
			}
			return map[string]any{"total": len(results), "boards": results}, nil
		},
	)
}

func (s *Skill) getSprintsTool() skill.Tool {
	return skill.NewTool(
		"jira_get_sprints",
		"List sprints for a Jira agile board",
		map[string]skill.Parameter{
			"board_id": {Type: "integer", Description: "Board ID to get sprints for", Required: true},
			"state":    {Type: "string", Description: "Filter by sprint state (active, closed, future)"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			boardID, ok := params["board_id"].(float64)
			if !ok {
				return nil, fmt.Errorf("board_id is required")
			}
			opts := &gojira.GetAllSprintsOptions{}
			if state, ok := params["state"].(string); ok {
				opts.State = state
			}
			sprints, err := s.client.BoardAPI.GetSprints(ctx, int(boardID), opts)
			if err != nil {
				return nil, fmt.Errorf("get sprints: %w", err)
			}
			results := make([]map[string]any, 0, len(sprints))
			for _, sp := range sprints {
				r := map[string]any{"id": sp.ID, "name": sp.Name, "state": sp.State}
				if sp.StartDate != "" {
					r["start_date"] = sp.StartDate
				}
				if sp.EndDate != "" {
					r["end_date"] = sp.EndDate
				}
				results = append(results, r)
			}
			return map[string]any{"board_id": int(boardID), "total": len(results), "sprints": results}, nil
		},
	)
}

func (s *Skill) velocityReportTool() skill.Tool {
	return skill.NewTool(
		"jira_velocity_report",
		"Get velocity report for a Jira agile board showing story points per sprint",
		map[string]skill.Parameter{
			"board_id":           {Type: "integer", Description: "Board ID to get velocity for", Required: true},
			"sprint_count":       {Type: "integer", Description: "Number of sprints to include"},
			"include_active":     {Type: "boolean", Description: "Include active sprint in calculations"},
			"story_points_field": {Type: "string", Description: "Custom field ID for story points (default: customfield_10016)"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			boardID, ok := params["board_id"].(float64)
			if !ok {
				return nil, fmt.Errorf("board_id is required")
			}
			opts := &jira.VelocityOptions{}
			if sc, ok := params["sprint_count"].(float64); ok {
				opts.SprintCount = int(sc)
			}
			if ia, ok := params["include_active"].(bool); ok {
				opts.IncludeActive = ia
			}
			if spf, ok := params["story_points_field"].(string); ok {
				opts.StoryPointsFieldID = spf
			}
			report, err := s.client.BoardAPI.GetVelocityReport(ctx, int(boardID), opts)
			if err != nil {
				return nil, fmt.Errorf("get velocity report: %w", err)
			}
			sprintData := make([]map[string]any, 0, len(report.Sprints))
			for _, sp := range report.Sprints {
				sprintData = append(sprintData, map[string]any{
					"sprint_id": sp.SprintID, "sprint_name": sp.SprintName,
					"completed_points": sp.CompletedPts, "issue_count": sp.IssueCount, "state": sp.State,
				})
			}
			return map[string]any{
				"board_id": int(boardID), "average_velocity": report.AveragePoints,
				"total_sprints": report.SprintCount, "total_points": report.TotalPoints, "sprints": sprintData,
			}, nil
		},
	)
}

func (s *Skill) burndownReportTool() skill.Tool {
	return skill.NewTool(
		"jira_burndown_report",
		"Get burndown report for a sprint showing remaining vs completed work",
		map[string]skill.Parameter{
			"sprint_id":          {Type: "integer", Description: "Sprint ID to get burndown for", Required: true},
			"story_points_field": {Type: "string", Description: "Custom field ID for story points (default: customfield_10016)"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			sprintID, ok := params["sprint_id"].(float64)
			if !ok {
				return nil, fmt.Errorf("sprint_id is required")
			}
			opts := &jira.BurndownOptions{}
			if spf, ok := params["story_points_field"].(string); ok {
				opts.StoryPointsFieldID = spf
			}
			report, err := s.client.BoardAPI.GetBurndownReport(ctx, int(sprintID), opts)
			if err != nil {
				return nil, fmt.Errorf("get burndown report: %w", err)
			}
			var completedPoints, remainingPoints float64
			var completedIssues, remainingIssues int
			if len(report.DailyData) > 0 {
				current := report.DailyData[len(report.DailyData)-1]
				completedPoints = current.CompletedPoints
				remainingPoints = current.RemainingPoints
				completedIssues = current.CompletedIssues
				remainingIssues = current.RemainingIssues
			}
			return map[string]any{
				"sprint_id": report.SprintID, "sprint_name": report.SprintName,
				"total_points": report.TotalPoints, "completed_points": completedPoints,
				"remaining_points": remainingPoints, "total_issues": report.TotalIssues,
				"completed_issues": completedIssues, "remaining_issues": remainingIssues,
				"start_date": report.StartDate, "end_date": report.EndDate,
			}, nil
		},
	)
}

func (s *Skill) worklogReportTool() skill.Tool {
	return skill.NewTool(
		"jira_worklog_report",
		"Get worklog summary report showing time spent by author and issue",
		map[string]skill.Parameter{
			"jql":       {Type: "string", Description: "JQL query to filter issues"},
			"sprint_id": {Type: "integer", Description: "Sprint ID to filter issues (alternative to jql)"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			opts := &jira.WorklogOptions{}
			if jql, ok := params["jql"].(string); ok {
				opts.JQL = jql
			}
			if sid, ok := params["sprint_id"].(float64); ok {
				opts.SprintID = int(sid)
			}
			if opts.JQL == "" && opts.SprintID == 0 {
				return nil, fmt.Errorf("either jql or sprint_id is required")
			}
			report, err := s.client.BoardAPI.GetWorklogReport(ctx, opts)
			if err != nil {
				return nil, fmt.Errorf("get worklog report: %w", err)
			}
			byAuthor := make([]map[string]any, 0, len(report.ByAuthor))
			for _, a := range report.ByAuthor {
				byAuthor = append(byAuthor, map[string]any{
					"author": a.Author, "time_spent": a.TimeSpentStr,
					"time_spent_secs": a.TimeSpent, "worklog_count": a.WorklogCount,
				})
			}
			byIssue := make([]map[string]any, 0, len(report.ByIssue))
			for _, i := range report.ByIssue {
				byIssue = append(byIssue, map[string]any{
					"key": i.Key, "summary": i.Summary, "time_spent": i.TimeSpentStr,
					"time_spent_secs": i.TimeSpent, "worklog_count": i.WorklogCount,
				})
			}
			return map[string]any{
				"total_time_spent": report.TotalTimeSpentStr, "total_time_spent_secs": report.TotalTimeSpent,
				"total_worklog_count": report.WorklogCount, "issue_count": report.IssueCount,
				"by_author": byAuthor, "by_issue": byIssue,
			}, nil
		},
	)
}

func (s *Skill) cycleTimeReportTool() skill.Tool {
	return skill.NewTool(
		"jira_cycle_time_report",
		"Get cycle time analysis for resolved issues showing time from creation to resolution",
		map[string]skill.Parameter{
			"jql":       {Type: "string", Description: "JQL query to filter issues"},
			"sprint_id": {Type: "integer", Description: "Sprint ID to filter issues (alternative to jql)"},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			opts := &jira.CycleTimeOptions{}
			if jql, ok := params["jql"].(string); ok {
				opts.JQL = jql
			}
			if sid, ok := params["sprint_id"].(float64); ok {
				opts.SprintID = int(sid)
			}
			if opts.JQL == "" && opts.SprintID == 0 {
				return nil, fmt.Errorf("either jql or sprint_id is required")
			}
			report, err := s.client.BoardAPI.GetCycleTimeReport(ctx, opts)
			if err != nil {
				return nil, fmt.Errorf("get cycle time report: %w", err)
			}
			issues := make([]map[string]any, 0, len(report.Issues))
			for _, i := range report.Issues {
				issues = append(issues, map[string]any{
					"key": i.Key, "summary": i.Summary, "type": i.IssueType,
					"created": i.Created, "resolved": i.Resolved, "cycle_time_days": i.CycleTimeDays,
				})
			}
			return map[string]any{
				"average_days": report.AverageCycleTime, "median_days": report.MedianCycleTime,
				"min_days": report.MinCycleTime, "max_days": report.MaxCycleTime,
				"total_resolved": report.IssueCount, "issues": issues,
			}, nil
		},
	)
}

func (s *Skill) moveToSprintTool() skill.Tool {
	return skill.NewTool(
		"jira_move_to_sprint",
		"Move issues to a sprint for sprint planning",
		map[string]skill.Parameter{
			"sprint_id":  {Type: "integer", Description: "Target sprint ID", Required: true},
			"issue_keys": {Type: "array", Description: "Issue keys to move to the sprint", Required: true},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			sprintID, ok := params["sprint_id"].(float64)
			if !ok {
				return nil, fmt.Errorf("sprint_id is required")
			}
			var issueKeys []string
			if keys, ok := params["issue_keys"].([]any); ok {
				for _, k := range keys {
					if ks, ok := k.(string); ok {
						issueKeys = append(issueKeys, ks)
					}
				}
			}
			if len(issueKeys) == 0 {
				return nil, fmt.Errorf("issue_keys is required")
			}
			result, err := s.client.BoardAPI.MoveIssuesToSprint(ctx, int(sprintID), issueKeys)
			if err != nil {
				return nil, fmt.Errorf("move to sprint: %w", err)
			}
			return map[string]any{
				"success": true, "sprint_id": result.SprintID,
				"issues_moved": result.IssuesMoved, "issue_count": result.IssueCount,
				"message": fmt.Sprintf("Moved %d issues to sprint %d", result.IssueCount, result.SprintID),
			}, nil
		},
	)
}
