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

// JobBoardExclusions contains -site: prefixed domains to exclude job boards and aggregators
var JobBoardExclusions = []string{
	// Major job boards
	"-site:linkedin.com", "-site:indeed.com", "-site:glassdoor.com",
	"-site:ziprecruiter.com", "-site:dice.com", "-site:monster.com",
	"-site:simplyhired.com", "-site:careerbuilder.com",
	// Tech job boards
	"-site:builtinnyc.com", "-site:builtin.com", "-site:jobright.ai",
	"-site:wellfound.com", "-site:roberthalf.com", "-site:hired.com",
	"-site:angel.co", "-site:stackoverflow.com/jobs",
	// Remote job boards
	"-site:remoteok.com", "-site:remoterocketship.com", "-site:weworkremotely.com",
	"-site:remote.co", "-site:flexjobs.com", "-site:dailyremote.com",
	// AI/ML specific job boards
	"-site:aijobs.com", "-site:aijobs.net", "-site:mljobs.com",
	// Aggregators
	"-site:jobleads.com", "-site:jobgether.com", "-site:jooble.org",
	"-site:neuvoo.com", "-site:talent.com", "-site:getwork.com",
	"-site:codingjobboard.com", "-site:snagajob.com", "-site:jobget.com",
	"-site:jobtarget.com", "-site:harnham.com", "-site:glocomms.com",
	"-site:hiringagents.com", "-site:funded.club",
}
