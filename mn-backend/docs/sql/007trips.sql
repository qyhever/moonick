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
      CASE WHEN seq.n % 2 = 1 THEN 'driver_post' ELSE 'passenger_post' END AS trip_type,
      CASE seq.n % 10
          WHEN 1 THEN '上海虹桥'
          WHEN 2 THEN '杭州东站'
          WHEN 3 THEN '苏州站'
          WHEN 4 THEN '南京南站'
          WHEN 5 THEN '无锡东站'
          WHEN 6 THEN '常州北站'
          WHEN 7 THEN '宁波站'
          WHEN 8 THEN '嘉兴南站'
          WHEN 9 THEN '绍兴北站'
          ELSE '湖州站'
      END AS from_text,
      CASE seq.n % 10
          WHEN 1 THEN '杭州东站'
          WHEN 2 THEN '上海虹桥'
          WHEN 3 THEN '南京南站'
          WHEN 4 THEN '苏州站'
          WHEN 5 THEN '上海南站'
          WHEN 6 THEN '无锡东站'
          WHEN 7 THEN '绍兴北站'
          WHEN 8 THEN '宁波站'
          WHEN 9 THEN '嘉兴南站'
          ELSE '苏州工业园区'
      END AS to_text,
      DATE_ADD('2026-05-10', INTERVAL seq.n - 1 DAY) AS departure_date,
      MAKETIME(6 + (seq.n % 12), (seq.n * 7) % 60, 0) AS departure_time,
      CASE
          WHEN seq.n % 3 = 0 THEN 1
          WHEN seq.n % 3 = 1 THEN 2
          ELSE 3
      END AS seat_count,
      CASE
          WHEN seq.n % 5 = 0 THEN 0.00
          ELSE 20 + seq.n
      END AS price_amount,
      CASE WHEN seq.n % 4 = 0 THEN 1 ELSE 0 END AS is_price_negotiable,
      CONCAT('active_trip_', LPAD(seq.n, 2, '0')) AS contact_wechat,
      CONCAT('1880001', LPAD(seq.n, 4, '0')) AS contact_phone,
      CONCAT(
          CASE WHEN seq.n % 2 = 1 THEN '车找人：' ELSE '人找车：' END,
          '测试行程 #', LPAD(seq.n, 2, '0'),
          '，可带简单行李，状态 active。'
      ) AS remark,
      'active' AS status,
      '' AS closed_reason,
      DATE_ADD('2026-05-01 08:00:00', INTERVAL seq.n HOUR) AS created_at,
      DATE_ADD('2026-05-01 08:00:00', INTERVAL seq.n HOUR) AS updated_at
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
  JOIN users u
      ON u.phone = CONCAT('1880001', LPAD(seq.n, 4, '0'))
  LEFT JOIN trips t
      ON t.publisher_user_id = u.id
     AND t.contact_wechat = CONCAT('active_trip_', LPAD(seq.n, 2, '0'))
     AND t.deleted_at IS NULL
  WHERE t.id IS NULL;

  COMMIT;