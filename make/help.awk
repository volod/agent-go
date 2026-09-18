# Prints "##@ Section" headings and "target: ## description" lines of the Makefiles.
BEGIN { print "Usage: make [VAR=value ...] <target>" }

/^##@ / { printf "\n%s:\n", substr($0, 5); next }

/^[A-Za-z0-9_.-]+:.*## / {
	target = $0
	sub(/:.*/, "", target)
	description = $0
	sub(/^[^#]*## */, "", description)
	printf "  %-16s %s\n", target, description
}
