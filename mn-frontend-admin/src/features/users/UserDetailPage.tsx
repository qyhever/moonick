import { useEffect, useState } from "react";
import { Card, Descriptions, Statistic, Table } from "antd";
import { useParams } from "react-router-dom";
import type { ColumnsType } from "antd/es/table";

import { getAdminUserDetail, getAdminUserTrips, type AdminUserDetail } from "./api";
import type { AdminTripSummary } from "../trips/api";

const columns: ColumnsType<AdminTripSummary> = [
  { title: "ID", dataIndex: "id", width: 90 },
  {
    title: "路线",
    render: (_, record) => `${record.fromText} → ${record.toText}`,
  },
  {
    title: "出发时间",
    render: (_, record) => `${record.departureDate} ${record.departureTime}`,
  },
  { title: "状态", dataIndex: "status" },
];

export default function UserDetailPage() {
  const { id = "" } = useParams();
  const [user, setUser] = useState<AdminUserDetail | null>(null);
  const [trips, setTrips] = useState<AdminTripSummary[]>([]);

  useEffect(() => {
    let active = true;

    async function load() {
      const [userDetail, tripList] = await Promise.all([
        getAdminUserDetail(id),
        getAdminUserTrips(id),
      ]);

      if (!active) {
        return;
      }

      setUser(userDetail);
      setTrips(tripList.items);
    }

    void load();
    return () => {
      active = false;
    };
  }, [id]);

  if (!user) {
    return <Card>加载中...</Card>;
  }

  return (
    <>
      <Card style={{ marginBottom: 16 }}>
        <Descriptions bordered title="基本资料">
          <Descriptions.Item label="昵称">{user.nickname}</Descriptions.Item>
          <Descriptions.Item label="手机号">{user.phone}</Descriptions.Item>
          <Descriptions.Item label="状态">{user.status}</Descriptions.Item>
          <Descriptions.Item label="默认微信">{user.defaultWechat || "未填写"}</Descriptions.Item>
          <Descriptions.Item label="默认手机号">{user.defaultPhone || "未填写"}</Descriptions.Item>
        </Descriptions>
      </Card>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(2, minmax(0, 1fr))", gap: 16, marginBottom: 16 }}>
        <Card>
          <Statistic title="发布数量" value={user.publishedTripCount} />
        </Card>
        <Card>
          <Statistic title="收藏数量" value={user.favoriteCount} />
        </Card>
      </div>

      <Card title="发布行程列表">
        <Table columns={columns} dataSource={trips} pagination={false} rowKey="id" />
      </Card>
    </>
  );
}
