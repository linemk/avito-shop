import http from 'k6/http';
import { sleep, check } from 'k6';

export let options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 1000,              // 1000 итераций (запросов) в секунду
            timeUnit: '1s',          // за одну секунду
            duration: '10s',          // длительность теста – 10 секунд
            preAllocatedVUs: 1000,    // заранее выделяем 1000 виртуальных пользователей
            maxVUs: 1000,            // максимум может быть до 1000 VUs при пиковых нагрузках
        },
    },
    thresholds: {
        // 95% запросов должны выполняться менее чем за 50 мс
        http_req_duration: ['p(95)<50'],
        // Доля неуспешных запросов (ошибок) должна быть меньше 0.01% (то есть успешность ≥ 99.99%)
        http_req_failed: ['rate<0.0001'],
    },
};

export default function () {
    // Если эндпоинт защищён JWT, не забудьте заменить <your_jwt_token> на валидный токен.
    let res = http.get('http://localhost:8080/api/info', {
        headers: {
            'Authorization': 'Bearer <your_jwt_valid>',
        },
    });

    check(res, {
        'status is 200': (r) => r.status === 200,
    });

    // Минимальная задержка, чтобы итерация не зацикливалась мгновенно.
    sleep(0.001);
}