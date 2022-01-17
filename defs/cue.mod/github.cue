package eventdefintions

pr: #Def & {
	integration: "github"
	identifiers: {
		"email": "The user's email"
	}
	schema: {
		name: "github.pull_request"

		data: {
			action: "opened" | "closed" | "merged" | "review_requested" | "synchronize" | "edited"

			// The pull request number.  Also contained within pull_request
			number: >=1

			pull_request: {
				// The pull request number.  Also specified within  top-level data
				number: >=1

				// The pull request title
				title: string

				// The pull request description
				body: string

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
