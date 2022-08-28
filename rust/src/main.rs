use axum::{extract::Extension, response::Html, response::IntoResponse, routing::get, Router};
use std::sync::Arc;
use tokio::sync::Mutex;

use evmap::{self, ReadHandleFactory, WriteHandle};
use rand;
use rand::Rng;
use std::collections::HashMap;
use tokio;

async fn request_handler(
    // Extension(rng): Extension<StdRng>,
    Extension(reader): Extension<ReadHandleFactory<i64, String>>,
    Extension(stats): Extension<Arc<Mutex<HashMap<String, i64>>>>,
) -> impl IntoResponse {
    let mut rng = rand::thread_rng();
    let n: i64 = rng.gen();
    let n = n.rem_euclid(49 as i64);
    let r = reader.handle();
    if let Some(res) = r.get(&n) {
        for re in res.as_ref() {
            let restr = re.to_string();
            let str_for_future = restr.clone();

            tokio::task::spawn(async move {
                let mut writer = stats.lock().await;
                *writer.entry(str_for_future).or_insert(0) += 1;
            });
            return Html(restr);
        }
    }
    return Html("error".to_string());
}
async fn admin_handler(stats: Extension<Arc<Mutex<HashMap<String, i64>>>>) -> Html<String> {
    let mut res = String::new();
    let data = stats.lock().await;
    let data = data.iter();
    for (key, value) in data {
        let r = format!("{} - {}\n", key, value);
        res.push_str(r.as_str())
    }
    Html(res)
}

async fn background_task(mut writer: WriteHandle<i64, String>) {
    let letters = "abcdefghijklmnopqrstuvwxyz".to_string();
    loop {
        tokio::time::sleep(tokio::time::Duration::from_millis(200)).await;
        let mut rng = rand::thread_rng();
        let i: i64 = rng.gen();
        let i = i.rem_euclid(50);
        let n: usize = rng.gen();
        let n = n.rem_euclid(letters.len());
        let n2: usize = rng.gen();
        let n2 = n2.rem_euclid(letters.len());
        let mut res = String::new();
        res.push(letters.as_bytes()[n] as char);
        res.push(letters.as_bytes()[n2] as char);
        writer.insert(i, res);
        writer.refresh();
    }
}

#[tokio::main]
async fn main() {
    let (reader, mut writer) = evmap::new();
    let letters = "abcdefghijklmnopqrstuvwxyz".to_string();
    let mut rng = rand::thread_rng();
    for i in 0..=50 {
        let n: usize = rng.gen();
        let n = n.rem_euclid(letters.len());
        let n2: usize = rng.gen();
        let n2 = n2.rem_euclid(letters.len());
        let mut res = String::new();
        res.push(letters.as_bytes()[n] as char);
        res.push(letters.as_bytes()[n2] as char);

        writer.insert(i as i64, res);
    }
    writer.refresh();
    let locked_stats = Arc::new(Mutex::new(HashMap::<String, i64>::new()));
    let read_factory = reader.factory();

    tokio::spawn(async move {
        background_task(writer).await;
    });
    let app = Router::new()
        .route("/request", get(request_handler))
        .route("/admin/requests", get(admin_handler))
        .layer(Extension(locked_stats))
        .layer(Extension(read_factory));

    axum::Server::bind(&"0.0.0.0:8081".parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}
