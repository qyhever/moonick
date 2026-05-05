-- 明叶同行县域行程测试数据
-- 用途：批量补充 50 条更贴近日常拼车场景的行程数据，支持重复执行
-- 依赖：需先导入 users 初始化数据（docs/sql/002_seed_users.sql）

START TRANSACTION;

INSERT INTO trips (
    publisher_user_id,
    trip_type,
    from_text,
    to_text,
    departure_date,
    departure_time,
    seat_count,
    price_amount,
    is_price_negotiable,
    contact_wechat,
    contact_phone,
    remark,
    status,
    closed_reason,
    created_at,
    updated_at
)
SELECT
    u.id AS publisher_user_id,
    seed.trip_type,
    seed.from_text,
    seed.to_text,
    seed.departure_date,
    seed.departure_time,
    seed.seat_count,
    seed.price_amount,
    seed.is_price_negotiable,
    CONCAT(u.default_wechat, '_ride', LPAD(seed.seq, 2, '0')) AS contact_wechat,
    CASE
        WHEN MOD(seed.seq, 3) = 0 THEN ''
        ELSE u.default_phone
    END AS contact_phone,
    seed.remark,
    seed.status,
    seed.closed_reason,
    seed.created_at,
    seed.updated_at
FROM (
    SELECT
        seq.n AS seq,
        CONCAT('1880002', LPAD(seq.n, 4, '0')) AS publisher_phone,
        CASE WHEN MOD(seq.n, 2) = 1 THEN 'driver_post' ELSE 'passenger_post' END AS trip_type,
        CASE MOD(seq.n - 1, 25)
            WHEN 0 THEN '昆山市'
            WHEN 1 THEN '吴江区'
            WHEN 2 THEN '张家港市'
            WHEN 3 THEN '江阴市'
            WHEN 4 THEN '溧阳市'
            WHEN 5 THEN '丹阳市'
            WHEN 6 THEN '义乌市'
            WHEN 7 THEN '诸暨市'
            WHEN 8 THEN '余姚市'
            WHEN 9 THEN '海宁市'
            WHEN 10 THEN '平湖市'
            WHEN 11 THEN '德清县'
            WHEN 12 THEN '长兴县'
            WHEN 13 THEN '乐清市'
            WHEN 14 THEN '平阳县'
            WHEN 15 THEN '温岭市'
            WHEN 16 THEN '临海市'
            WHEN 17 THEN '晋江市'
            WHEN 18 THEN '石狮市'
            WHEN 19 THEN '龙海区'
            WHEN 20 THEN '福清市'
            WHEN 21 THEN '三河市'
            WHEN 22 THEN '固安县'
            WHEN 23 THEN '琼海市'
            ELSE '文昌市'
        END AS from_text,
        CASE MOD(seq.n - 1, 25)
            WHEN 0 THEN '太仓市'
            WHEN 1 THEN '常熟市'
            WHEN 2 THEN '太仓市'
            WHEN 3 THEN '宜兴市'
            WHEN 4 THEN '金坛区'
            WHEN 5 THEN '句容市'
            WHEN 6 THEN '东阳市'
            WHEN 7 THEN '嵊州市'
            WHEN 8 THEN '慈溪市'
            WHEN 9 THEN '桐乡市'
            WHEN 10 THEN '嘉善县'
            WHEN 11 THEN '安吉县'
            WHEN 12 THEN '宜兴市'
            WHEN 13 THEN '瑞安市'
            WHEN 14 THEN '苍南县'
            WHEN 15 THEN '玉环市'
            WHEN 16 THEN '仙居县'
            WHEN 17 THEN '南安市'
            WHEN 18 THEN '晋江市'
            WHEN 19 THEN '长泰区'
            WHEN 20 THEN '长乐区'
            WHEN 21 THEN '香河县'
            WHEN 22 THEN '涿州市'
            WHEN 23 THEN '万宁市'
            ELSE '琼海市'
        END AS to_text,
        DATE_ADD('2026-05-06', INTERVAL seq.n - 1 DAY) AS departure_date,
        CASE MOD(seq.n - 1, 10)
            WHEN 0 THEN '07:10:00'
            WHEN 1 THEN '08:20:00'
            WHEN 2 THEN '09:30:00'
            WHEN 3 THEN '10:40:00'
            WHEN 4 THEN '11:50:00'
            WHEN 5 THEN '13:00:00'
            WHEN 6 THEN '14:10:00'
            WHEN 7 THEN '15:20:00'
            WHEN 8 THEN '16:30:00'
            ELSE '17:40:00'
        END AS departure_time,
        CASE
            WHEN MOD(seq.n, 2) = 1 THEN (MOD(seq.n, 4) + 1)
            ELSE (MOD(seq.n, 3) + 1)
        END AS seat_count,
        CASE
            WHEN MOD(seq.n, 7) = 0 THEN 0.00
            ELSE 28 + seq.n * 3 + MOD(seq.n, 4) * 0.50
        END AS price_amount,
        CASE WHEN MOD(seq.n, 4) = 0 THEN 1 ELSE 0 END AS is_price_negotiable,
        CASE
            WHEN MOD(seq.n, 2) = 1 THEN CONCAT(
                '车找人：',
                CASE MOD(seq.n - 1, 25)
                    WHEN 0 THEN '昆山市'
                    WHEN 1 THEN '吴江区'
                    WHEN 2 THEN '张家港市'
                    WHEN 3 THEN '江阴市'
                    WHEN 4 THEN '溧阳市'
                    WHEN 5 THEN '丹阳市'
                    WHEN 6 THEN '义乌市'
                    WHEN 7 THEN '诸暨市'
                    WHEN 8 THEN '余姚市'
                    WHEN 9 THEN '海宁市'
                    WHEN 10 THEN '平湖市'
                    WHEN 11 THEN '德清县'
                    WHEN 12 THEN '长兴县'
                    WHEN 13 THEN '乐清市'
                    WHEN 14 THEN '平阳县'
                    WHEN 15 THEN '温岭市'
                    WHEN 16 THEN '临海市'
                    WHEN 17 THEN '晋江市'
                    WHEN 18 THEN '石狮市'
                    WHEN 19 THEN '龙海区'
                    WHEN 20 THEN '福清市'
                    WHEN 21 THEN '三河市'
                    WHEN 22 THEN '固安县'
                    WHEN 23 THEN '琼海市'
                    ELSE '文昌市'
                END,
                '到',
                CASE MOD(seq.n - 1, 25)
                    WHEN 0 THEN '太仓市'
                    WHEN 1 THEN '常熟市'
                    WHEN 2 THEN '太仓市'
                    WHEN 3 THEN '宜兴市'
                    WHEN 4 THEN '金坛区'
                    WHEN 5 THEN '句容市'
                    WHEN 6 THEN '东阳市'
                    WHEN 7 THEN '嵊州市'
                    WHEN 8 THEN '慈溪市'
                    WHEN 9 THEN '桐乡市'
                    WHEN 10 THEN '嘉善县'
                    WHEN 11 THEN '安吉县'
                    WHEN 12 THEN '宜兴市'
                    WHEN 13 THEN '瑞安市'
                    WHEN 14 THEN '苍南县'
                    WHEN 15 THEN '玉环市'
                    WHEN 16 THEN '仙居县'
                    WHEN 17 THEN '南安市'
                    WHEN 18 THEN '晋江市'
                    WHEN 19 THEN '长泰区'
                    WHEN 20 THEN '长乐区'
                    WHEN 21 THEN '香河县'
                    WHEN 22 THEN '涿州市'
                    WHEN 23 THEN '万宁市'
                    ELSE '琼海市'
                END,
                '，可带1件行李，守时出发。'
            )
            ELSE CONCAT(
                '人找车：',
                CASE MOD(seq.n - 1, 25)
                    WHEN 0 THEN '昆山市'
                    WHEN 1 THEN '吴江区'
                    WHEN 2 THEN '张家港市'
                    WHEN 3 THEN '江阴市'
                    WHEN 4 THEN '溧阳市'
                    WHEN 5 THEN '丹阳市'
                    WHEN 6 THEN '义乌市'
                    WHEN 7 THEN '诸暨市'
                    WHEN 8 THEN '余姚市'
                    WHEN 9 THEN '海宁市'
                    WHEN 10 THEN '平湖市'
                    WHEN 11 THEN '德清县'
                    WHEN 12 THEN '长兴县'
                    WHEN 13 THEN '乐清市'
                    WHEN 14 THEN '平阳县'
                    WHEN 15 THEN '温岭市'
                    WHEN 16 THEN '临海市'
                    WHEN 17 THEN '晋江市'
                    WHEN 18 THEN '石狮市'
                    WHEN 19 THEN '龙海区'
                    WHEN 20 THEN '福清市'
                    WHEN 21 THEN '三河市'
                    WHEN 22 THEN '固安县'
                    WHEN 23 THEN '琼海市'
                    ELSE '文昌市'
                END,
                '到',
                CASE MOD(seq.n - 1, 25)
                    WHEN 0 THEN '太仓市'
                    WHEN 1 THEN '常熟市'
                    WHEN 2 THEN '太仓市'
                    WHEN 3 THEN '宜兴市'
                    WHEN 4 THEN '金坛区'
                    WHEN 5 THEN '句容市'
                    WHEN 6 THEN '东阳市'
                    WHEN 7 THEN '嵊州市'
                    WHEN 8 THEN '慈溪市'
                    WHEN 9 THEN '桐乡市'
                    WHEN 10 THEN '嘉善县'
                    WHEN 11 THEN '安吉县'
                    WHEN 12 THEN '宜兴市'
                    WHEN 13 THEN '瑞安市'
                    WHEN 14 THEN '苍南县'
                    WHEN 15 THEN '玉环市'
                    WHEN 16 THEN '仙居县'
                    WHEN 17 THEN '南安市'
                    WHEN 18 THEN '晋江市'
                    WHEN 19 THEN '长泰区'
                    WHEN 20 THEN '长乐区'
                    WHEN 21 THEN '香河县'
                    WHEN 22 THEN '涿州市'
                    WHEN 23 THEN '万宁市'
                    ELSE '琼海市'
                END,
                '，希望顺路拼车，时间可小幅协调。'
            )
        END AS remark,
        CASE
            WHEN seq.n <= 40 THEN 'active'
            WHEN seq.n <= 46 THEN 'full'
            ELSE 'closed'
        END AS status,
        CASE
            WHEN seq.n >= 47 THEN '行程已取消，改期再约'
            ELSE ''
        END AS closed_reason,
        DATE_ADD(
            DATE_ADD('2026-05-01 09:00:00', INTERVAL MOD(seq.n - 1, 5) DAY),
            INTERVAL seq.n MINUTE
        ) AS created_at,
        DATE_ADD(
            DATE_ADD('2026-05-01 09:00:00', INTERVAL MOD(seq.n - 1, 5) DAY),
            INTERVAL seq.n MINUTE
        ) AS updated_at
    FROM (
        SELECT 1 AS n UNION ALL
        SELECT 2 UNION ALL
        SELECT 3 UNION ALL
        SELECT 4 UNION ALL
        SELECT 5 UNION ALL
        SELECT 6 UNION ALL
        SELECT 7 UNION ALL
        SELECT 8 UNION ALL
        SELECT 9 UNION ALL
        SELECT 10 UNION ALL
        SELECT 11 UNION ALL
        SELECT 12 UNION ALL
        SELECT 13 UNION ALL
        SELECT 14 UNION ALL
        SELECT 15 UNION ALL
        SELECT 16 UNION ALL
        SELECT 17 UNION ALL
        SELECT 18 UNION ALL
        SELECT 19 UNION ALL
        SELECT 20 UNION ALL
        SELECT 21 UNION ALL
        SELECT 22 UNION ALL
        SELECT 23 UNION ALL
        SELECT 24 UNION ALL
        SELECT 25 UNION ALL
        SELECT 26 UNION ALL
        SELECT 27 UNION ALL
        SELECT 28 UNION ALL
        SELECT 29 UNION ALL
        SELECT 30 UNION ALL
        SELECT 31 UNION ALL
        SELECT 32 UNION ALL
        SELECT 33 UNION ALL
        SELECT 34 UNION ALL
        SELECT 35 UNION ALL
        SELECT 36 UNION ALL
        SELECT 37 UNION ALL
        SELECT 38 UNION ALL
        SELECT 39 UNION ALL
        SELECT 40 UNION ALL
        SELECT 41 UNION ALL
        SELECT 42 UNION ALL
        SELECT 43 UNION ALL
        SELECT 44 UNION ALL
        SELECT 45 UNION ALL
        SELECT 46 UNION ALL
        SELECT 47 UNION ALL
        SELECT 48 UNION ALL
        SELECT 49 UNION ALL
        SELECT 50
    ) AS seq
) AS seed
JOIN users u
    ON u.phone = seed.publisher_phone
LEFT JOIN trips t
    ON t.publisher_user_id = u.id
   AND t.contact_wechat = CONCAT(u.default_wechat, '_ride', LPAD(seed.seq, 2, '0'))
   AND t.deleted_at IS NULL
WHERE t.id IS NULL;

COMMIT;
