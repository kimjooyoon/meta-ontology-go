package main

import "flag"

type options struct {
	before, semanticAfter, commentAfter, metaAfter, corpus, out string
}

func parseOptions(args []string) (options, bool) {
	flags := flag.NewFlagSet("counterexample-first-intervention-witness", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.before, "before", "", "baseline Gooo source")
	flags.StringVar(&value.semanticAfter, "semantic-after", "", "semantically changed Gooo source")
	flags.StringVar(&value.commentAfter, "comment-after", "", "comment-only Gooo source")
	flags.StringVar(&value.metaAfter, "meta-after", "", "meta-operation graph intervention source")
	flags.StringVar(&value.corpus, "corpus", "", "fixed corpus inputs")
	flags.StringVar(&value.out, "out", "", "intervention artifact")
	valid := flags.Parse(args) == nil && value.before != "" && value.semanticAfter != "" && value.commentAfter != "" && value.metaAfter != "" && value.corpus != "" && value.out != ""
	return value, valid
}
