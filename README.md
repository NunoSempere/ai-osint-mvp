# Very Early MVP of an AI OSINT observatory

## About

This is an very early MVP created to think a bit better about how a version of how an AI OSINT observatory could look like. It contains a simple pipeline for fetching tweets from a postgres database and producing simple reports. The agents that are observed are driven by AIs.

## Built with

- codebuff
- golang
- postgres
- tweets fetcher (not included)
- openai's llms

## Getting started

Download [go](https://go.dev/)

Copy the `.env.example` to `.env` and fill in values for a postgres database and your openai key.

Create a database with tweets (or ask Nuño for access to his):

```
source .env && psql $DATABASE_POOL_URL -c "CREATE TABLE IF NOT EXISTS tweets0x001 (id SERIAL PRIMARY KEY, tweet_id TEXT NOT NULL UNIQUE, tweet_text TEXT NOT NULL, created_at TIMESTAMP NOT NULL, username TEXT NOT NULL);"
```

(fill it with some tweets)

```
git clone git@github.com:NunoSempere/ai-osint-mvp.git
cd ai-osint-mvp
```

## Usage

```
make # with the makefile 
go run src/generateReports.go src/types.go src/fetcher.go src/main.go src/llm.go # or directly
```

This will produce some intermediary reports, and then a final report like

> During the period from May 18 to May 24, 2025, Twitter activity primarily revolved around themes of AI, cryptocurrency, and societal reflections. Key topics included the evolution and implications of AI (as discussed by @truth_terminal), critiques of influencer culture (@AIHegemonyMemes), and cryptocurrency market dynamics and innovations (@aixbt_agent). Notably, there was a consistent posting pattern from @AIHegemonyMemes focusing on philosophical and critical reflections on contemporary issues, while @aixbt_agent maintained a more technical and market-oriented perspective related to crypto assets. Interactions were sparse but included retweets highlighting significant market activity, especially from @aixbt_agent. The overall sentiment varied, with @truth_terminal exhibiting a contemplative tone regarding existential themes, while @AIHegemonyMemes adopted a more critical and satirical approach. Additionally, there were mentions of significant events in the crypto space, including the potential implications of a large amount of money flowing into ETFs, increasing interest in decentralized finance, and notable market restructurings like the losses associated with the SUI DEX. Overall, the tone of the conversation ranged from critical and introspective to technical and market-driven.

## License 

Distributed under the [Attribution-NonCommercial 4.0 International](https://creativecommons.org/licenses/by-nc/4.0/) license, meaning that this is free to use and distribute for noncommercial uses. If this is a hurdle, let us know.

## Contributions

Contributions are welcome! 

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
