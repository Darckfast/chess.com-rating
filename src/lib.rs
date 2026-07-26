use std::time::Duration;

use http::StatusCode;
use reqwest::Client;
use serde::{Deserialize, de::IgnoredAny};
use worker::{Context, Env, Request, Response, console_log, console_warn, event, send::SendFuture};

static STATS_KEYS: [&str; 6] = [
    "chess",
    "lightning",
    "tactics",
    "rapid",
    "tactics_challenge",
    "bullet",
];

#[event(fetch)]
async fn fetch(req: Request, env: Env, _ctx: Context) -> worker::Result<Response> {
    match env.durable_object("RUSTY_LIMITER") {
        Ok(stub) => {
            let headers = req.headers();
            let id = headers
                .get("cf-connecting-ip")
                .or(headers.get("x-forwarded-for"))
                .unwrap_or(Some("127.0.0.1".to_string()))
                .unwrap();

            let limiter = stub.get_by_name(&id)?;
            let rs = SendFuture::new(async move {
                limiter
                    .fetch_with_str(&"http://rate-limit".to_string())
                    .await
            })
            .await?;

            if rs.status_code() == http::StatusCode::TOO_MANY_REQUESTS {
                return Response::error("", 429);
            }
        }
        Err(_) => {
            console_warn!("rate limit not enabled")
        }
    }

    let params = if let Ok(params) = req.query::<Qs>() {
        params
    } else {
        return Response::ok("username and message are required");
    };

    let client = Client::builder()
        .timeout(Duration::from_secs(5))
        .build()
        .map_err(|e| worker::Error::InternalError(e.to_string()))?;

    match client
        .get(format!(
            "https://www.chess.com/callback/member/stats/{}",
            params.username
        ))
        .send()
        .await
    {
        Ok(rs) => {
            let status = rs.status();
            match status {
                StatusCode::OK => {
                    let j = rs.json::<Chess>().await.unwrap();

                    let mut message = params.message;
                    for k in STATS_KEYS {
                        let s = j.stats.iter().find(|p| p.key == k.to_string());

                        match s {
                            Some(ss) => {
                                message =
                                    message.replace(&format!("={k}"), &ss.stats.rating.to_string());
                            }
                            None => (),
                        }
                    }

                    Response::ok(message)
                }
                StatusCode::NOT_FOUND => {
                    console_log!("User not found on chess.com {}", params.username);
                    Response::ok(format!("user {} not found", params.username))
                }
                _ => {
                    console_log!("Error fetching user on chess.com {}", status);
                    Response::ok("error fetching user on chess.com")
                }
            }
        }
        Err(err) => {
            if err.is_timeout() {
                Response::ok("chess.com took too long to respond :(")
            } else {
                console_log!("Error fetching user on chess.com {}", err);
                Response::error("error fetching user on chess.com", 500)
            }
        }
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
struct Qs {
    username: String,
    message: String,
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
