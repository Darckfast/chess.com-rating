use serde::{de::IgnoredAny, Deserialize};
use std::collections::HashMap;
use worker::*;

static STATS_KEYS: [&str; 6] = [
    "chess",
    "lightning",
    "tactics",
    "rapid",
    "tactics_challenge",
    "bullet",
];

#[event(fetch)]
async fn fetch(_req: HttpRequest, _env: Env, _ctx: Context) -> Result<Response> {
    let params: HashMap<String, String> = _req
        .uri()
        .query()
        .map(|a| {
            url::form_urlencoded::parse(a.as_bytes())
                .into_owned()
                .collect()
        })
        .unwrap_or_else(HashMap::new);

    let username = params.get("username");
    let msg_tmpl = params.get("message");

    match username {
        Some(usr) => {
            let resp =
                reqwest::get("https://www.chess.com/callback/member/stats/".to_string() + usr)
                    .await
                    .unwrap()
                    .json::<Chess>()
                    .await
                    .unwrap();

            let mut rawstats: HashMap<String, String> = HashMap::new();

            for k in &STATS_KEYS {
                let s = resp.stats.iter().find(|p| p.key == k.to_string());

                match s {
                    Some(ss) => {
                        rawstats.insert(k.to_string(), ss.stats.rating.to_string());
                    }
                    None => {}
                }
            }

            if let Some(msg) = msg_tmpl {
                let message = msg
                    .split_whitespace()
                    .map(|word| match word {
                        _ if word.contains("=lightning") => rawstats
                            .get("lightning")
                            .map_or("not found", String::as_str)
                            .to_owned(),
                        _ if word.contains("=chess") => rawstats
                            .get("chess")
                            .map_or("not found", String::as_str)
                            .to_owned(),
                        _ if word.contains("=bullet") => rawstats
                            .get("bullet")
                            .map_or("not found", String::as_str)
                            .to_owned(),
                        _ if word.contains("=rapid") => rawstats
                            .get("rapid")
                            .map_or("not found", String::as_str)
                            .to_owned(),
                        _ if word.contains("=tactics_challenge") => rawstats
                            .get("tactics_challenge")
                            .map_or("not found", String::as_str)
                            .to_owned(),
                        _ if word.contains("=tactics") => rawstats
                            .get("tactics")
                            .map_or("not found", String::as_str)
                            .to_owned(),
                        _ => word.to_owned(),
                    })
                    .collect::<Vec<String>>()
                    .join(" ");

                return Response::ok(message);
            } else {
                return Response::ok("message is required");
            }
        }
        None => return Response::ok("username is required"),
    }
}

fn default_skipped_int() -> i32 {
    0
}

fn deserialize_int_or_skip<'de, D>(deserializer: D) -> Result<i32, D::Error>
where
    D: serde::Deserializer<'de>,
{
    #[derive(Deserialize)]
    #[serde(untagged)]
    enum IntOrAny {
        Int(i32),
        Any(IgnoredAny),
    }

    let val = IntOrAny::deserialize(deserializer)?;
    match val {
        IntOrAny::Int(i) => Ok(i),
        IntOrAny::Any(_) => Ok(default_skipped_int()),
    }
}

#[derive(Deserialize, Debug)]
struct Rating {
    #[serde(default = "default_skipped_int")]
    #[serde(deserialize_with = "deserialize_int_or_skip")]
    rating: i32,
}

#[derive(Deserialize, Debug)]
struct Destats {
    key: String,
    stats: Rating,
}

#[derive(Deserialize, Debug)]
struct Chess {
    stats: Vec<Destats>,
}
