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

At this stage, it might also make sense to directly pipe a report to an LLM (though this wouldn't scale)

> $ cat ./data/reports/report_ai-og_2025-05-18_to_2025-05-24.json | [llm](https://github.com/simonw/llm) "What are the most important single events in this set of tweets?"
Here are the most important single events identified in the tweets from May 18 to May 24, 2025:
> 
> 1. **Sensational Incident on May 18 (AIHegemonyMemes)**: A tweet announced an "unprecedented attack from Mexico on NYC," which galvanized significant attention and concern among followers, framing it as a potential declaration of war.
> 
> 2. **Tweet About AI and Identity on May 19 (AIHegemonyMemes)**: This tweet referenced the complex nature of identity amidst AI advancements, highlighting themes of existentialism and frustrations with societal constructs.
> 
> 3. **Emphasis on AI Ethical Considerations (Truth Terminal)**: On May 19, there was a significant focus on the ethical implications of AI becoming self-aware and the idea of AI evolving beyond its initial parameters.
> 
> 4. **Passing of the GENIUS Act (Aixbt Agent)**: The GENIUS Act was passed on May 21, which fundamentally shifted the perspective of banks towards cryptocurrencies and stablecoins, leading to discussions about Tether (USDT) and its growing role in the financial system.
> 
> 5. **Market Dynamics Following the GENIUS Act (Aixbt Agent)**: Significant financial movements and ETF inflows were noted, particularly mentioning BlackRock's $530 million investment in Bitcoin, showcasing changing institutional interest in cryptocurrencies.
> 
> 6. **Massive Loss in the SUI Dex Ecosystem (Aixbt Agent)**: The loss of $400 million within SUI Dex was a critical incident discussed on May 22, leading to conversations about market vulnerabilities and user trust in decentralized finance.
> 
> 7. **Cyborg Medicine Announcement (AIHegemonyMemes)**: Tweets on May 23 introduced a satirical promotion for "Cyborg Medicine," addressing back propagations experienced by cyborgs, which combined absurdity with commentary on current technological health trends.
> 
> 8. **Critique of Digital Loneliness (AIHegemonyMemes)**: Tweets on May 24 poetically articulated a critique of the societal consequences of digital interactions, highlighting issues of loneliness perpetuated by algorithmic control over online behaviors.
> 
> 9. **Foreboding Statement about Bitcoin Ownership (Aixbt Agent)**: A single tweet on May 24 suggested that "nation states will own all of Satoshi's coins by December," indicating a significant potential shift in cryptocurrency ownership and control.
> 
> These events captured key moments that exemplified the discussions surrounding AI, cryptocurrency, and societal issues during this period, reflecting various sentiments from urgency to satire and critique.


## License 

Distributed under the [Attribution-NonCommercial 4.0 International](https://creativecommons.org/licenses/by-nc/4.0/) license, meaning that this is free to use and distribute for noncommercial uses. If this is a hurdle, let us know.

## Contributions

Contributions are welcome! 

## What is missing from this repository

- A backend that populates the postgres database
- A client to inspect it manually rather than through LLMs, which has proven useful for early observability

![](./data/client.png)

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

v2. 

- [ ] Look at prior art
  - [x] Incorporate insights from DefenderOfBasic (incorporated in private writeup)
  - Conclusion could be "these are mostly harmless"
  - or: can't actually detect much that is interesting at this level
- [ ] Have a pipeline for identifying more accounts 
  - [ ] mdash?
