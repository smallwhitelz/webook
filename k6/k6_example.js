import http from 'k6/http';

const url = 'http://localhost:8080/hello';

export default function () {
    const data = { name: 'Bert' };

    const cb200 = http.expectedStatuses(200)
    // Using a JSON string as body
    const res = http.post(url, JSON.stringify(data), {
        headers: { 'Content-Type': 'application/json' },
        expectedStatuses: cb200,
    });
}