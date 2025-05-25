# Very Early MVP of an AI OSINT observatory

## About

This is an early MVP created to think a bit better about how an AI OSINT observatory might look like. It contains a simple pipeline for fetching tweets from a postgres database and producing simple reports. The agents that are observed are driven by AIs.

## Built with

- codebuff
- golang
- postgres
- tweets fetcher (not included)
- openai's llms

## Roadmap

v0. Save tweets

- [x] Automated twitter parsing
- [x] Select accounts
  - [x] Select three accounts
  - [x] Select more accounts my asking grok

v1. Create pipeline

- [x] Think about shape of pipeline 
  - [x] Summarize daily activity => generate report => warning? 
- [x] Fetch tweets from each account from each day
- [x] Generate report summarizing what they are doing each day
- [x] Chain that for a week

v2. Draft report

- [ ] Look at prior art
  - [ ] Incorporate insights from DefenderOfBasic
  - Conclusion could be "these are mostly harmless"
  - or: can't actually detect much that is interesting at this level
- [ ] Have a pipeline for identifying more accounts 
  - [ ] mdash?
