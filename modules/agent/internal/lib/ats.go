package lib

// ATSSites contains site: prefixed domains for ATS platforms where companies post directly
var ATSSites = []string{
	"site:greenhouse.io",
	"site:boards.greenhouse.io",
	"site:lever.co",
	"site:jobs.ashbyhq.com",
	"site:apply.workable.com",
	"site:jobs.smartrecruiters.com",
	"site:jobs.jobvite.com",
	"site:*.breezy.hr",
	"site:*.recruitee.com",
	"site:jobs.icims.com",
	"site:*.pinpointhq.com",
}

// ATSDomains contains domain patterns for identifying ATS URLs (without site: prefix)
var ATSDomains = []string{
	"greenhouse.io",
	"lever.co",
	"ashbyhq.com",
	"workable.com",
	"smartrecruiters.com",
	"jobvite.com",
	"breezy.hr",
	"recruitee.com",
	"icims.com",
	"pinpointhq.com",
}
