import { useEffect, useState } from "react";
import { Card, Table, Tag } from "antd";
import { Link } from "react-router-dom";
import type { ColumnsType } from "antd/es/table";

import TripSearchForm from "./components/TripSearchForm";
import { getAdminTrips, type AdminTripQuery, type AdminTripSummary } from "./api";

const statusColorMap: Record<string, string> = {
  active: "blue",
  full: "gold",
  closed: "default",
  expired: "red",
};

const columns: ColumnsType<AdminTripSummary> = [
  { title: "ID", dataIndex: "id", width: 90 },
  { title: "类型", dataIndex: "tripType" },
  {
    title: "路线",
    render: (_, record) => `${record.fromText} → ${record.toText}`,
  },
  {
    title: "出发时间",
    render: (_, record) => `${record.departureDate} ${record.departureTime}`,
  },
  { title: "人数", dataIndex: "seatCount", width: 90 },
  {
    title: "状态",
    dataIndex: "status",
    render: (value: string) => <Tag color={statusColorMap[value] || "default"}>{value}</Tag>,
  },
  {
    title: "操作",
    render: (_, record) => (
      <>
        <Link to={`/trips/${record.id}`}>详情</Link>
        <span style={{ margin: "0 8px" }}>|</span>
        <Link to={`/trips/${record.id}/edit`}>编辑</Link>
      </>
    ),
  },
];

export default function TripListPage() {
  const [data, setData] = useState<AdminTripSummary[]>([]);
  const [loading, setLoading] = useState(false);

  async function loadTrips(query: AdminTripQuery = {}) {
    setLoading(true);
    try {
      const response = await getAdminTrips(query);
      setData(response.items);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadTrips();
  }, []);

  return (
    <Card title="行程管理">
      <TripSearchForm onSearch={(values) => void loadTrips(values)} />
      <Table columns={columns} dataSource={data} loading={loading} pagination={false} rowKey="id" />
    </Card>
  );
}
