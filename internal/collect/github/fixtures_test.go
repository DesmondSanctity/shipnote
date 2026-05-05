package github_test

const happyPathResponse = `{"data":{"repository":{
"c0": {
	"oid":"aaa",
	"associatedPullRequests":{"nodes":[{
		"number":42,"title":"feat: x","body":"hello","url":"https://gh/pr/42",
		"state":"MERGED","mergedAt":"2026-04-22T10:14:00Z",
		"baseRefName":"main","headRefName":"feature",
		"mergeCommit":{"oid":"merge1"},
		"author":{"login":"maya","name":"Maya R.","url":"https://gh/maya","avatarUrl":"a","__typename":"User"},
		"labels":{"nodes":[{"name":"feat","color":"00ff00"}]},
		"files":{"nodes":[{"path":"src/x.go","additions":10,"deletions":2,"changeType":"MODIFIED"}]}
	}]}
},
"c1": {"oid":"bbb","associatedPullRequests":{"nodes":[]}}
}}}`
