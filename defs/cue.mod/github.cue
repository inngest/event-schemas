package eventdefintions

pr: #Def & {
	integration: "github"
	identifiers: {
		"email": "The user's email"
	}
	schema: {
		name: "github.com.pull_request"
		data: {
			action: "opened" | "closed" | "merged" | "review_requested" | "synchronize" | "edited"
			number: >=1

			pull_request: {
				// The pull request title
				title: string

				// The number of individual commits wanting to be merged
				commits: >=1

				// The number of changed files
				changed_files: >=1

				// Whether the pull request is a draft
				draft: bool
			}

			// The commit hash of the tip of the PR before changes
			before?: string
			// The commit hash of the tip of the PR after changes
			after?: string
		}
	}
}
