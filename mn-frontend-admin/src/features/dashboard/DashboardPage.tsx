import { useEffect, useState } from "react";
import { Card, Col, Row, Space, Statistic, Typography } from "antd";

import { getDashboardSummary, type DashboardSummary } from "./api";

const initialSummary: DashboardSummary = {
  totalUsers: 0,
  totalTrips: 0,
  activeTrips: 0,
  expiredTrips: 0,
  totalFavorites: 0,
};

export default function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary>(initialSummary);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;

    async function loadSummary() {
      try {
        const data = await getDashboardSummary();
        if (active) {
          setSummary(data);
        }
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "看板加载失败");
        }
      }
    }

    void loadSummary();

    return () => {
      active = false;
    };
  }, []);

  return (
    <Space direction="vertical" size={20} style={{ width: "100%" }}>
      <div>
        <Typography.Text type="secondary">轻量看板</Typography.Text>
        <Typography.Title level={2} style={{ marginTop: 4 }}>
          运营概览
        </Typography.Title>
      </div>

      {error ? <Card>{error}</Card> : null}

      <Row gutter={[16, 16]}>
        <Col span={12}>
          <Card>
            <Statistic title="行程总数" value={summary.totalTrips} />
          </Card>
        </Col>
        <Col span={12}>
          <Card>
            <Statistic title="用户总数" value={summary.totalUsers} />
          </Card>
        </Col>
        <Col span={12}>
          <Card>
            <Statistic title="当前有效行程数" value={summary.activeTrips} />
          </Card>
        </Col>
        <Col span={12}>
          <Card>
            <Statistic title="过期行程数" value={summary.expiredTrips} />
          </Card>
        </Col>
      </Row>

      <Card>
        <Statistic title="收藏总数" value={summary.totalFavorites} />
      </Card>
    </Space>
  );
}
