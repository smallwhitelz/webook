
import http from 'k6/http';
// http.setTimeout(30000);
export let options = {
   duration: '10s',
    vus: 10,
    rpc: 10,
};
export default  () => {
    var url = "http://127.0.0.1:8088/feed/list";
    var payload = JSON.stringify({
        uid: 30001,
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
