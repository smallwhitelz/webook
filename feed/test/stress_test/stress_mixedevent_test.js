import http from 'k6/http';
// http.setTimeout(30000);
export let options = {
    // 执行时间
    duration: '10s',
    // 并发量
    vus: 100,
    // 每秒多少请求
    rpc: 20,
};
export default  () => {
    var url = "http://127.0.0.1:8088/feed/list";
    var payload = JSON.stringify({
        uid: 40001,
        limit: 10,
        timestamp: 1708748101,
    });

    var params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    http.post(url, payload, params);
};