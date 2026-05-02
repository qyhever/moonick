import { useEffect, useState } from "react";
import { Card, Table, Tag } from "antd";
import { Link } from "react-router-dom";
import type { ColumnsType } from "antd/es/table";
import type { TablePaginationConfig } from "antd/es/table";

import TripSearchForm from "./components/TripSearchForm";
import { getAdminTrips, type AdminTripQuery, type AdminTripSummary } from "./api";
import { getTripStatusText, getTripTypeText } from "../displayText";

const statusColorMap: Record<string, string> = {
  active: "blue",
  full: "gold",
  closed: "default",
  expired: "red",
};

const columns: ColumnsType<AdminTripSummary> = [
  { title: "ID", dataIndex: "id", width: 90 },
  {
    title: "类型",
    dataIndex: "tripType",
    render: (value: string) => getTripTypeText(value),
  },
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
    render: (value: string) => <Tag color={statusColorMap[value] || "default"}>{getTripStatusText(value)}</Tag>,
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
  const [query, setQuery] = useState<AdminTripQuery>({});
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  async function loadTrips(nextQuery: AdminTripQuery = {}, nextPageNum = 1, nextPageSize = pagination.pageSize) {
    setLoading(true);
    try {
      const response = await getAdminTrips({
        ...nextQuery,
        pageNum: nextPageNum,
        pageSize: nextPageSize,
      });
      setData(response.items);
      setPagination({
        current: response.pageNum,
        pageSize: response.pageSize,
        total: response.total,
      });
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadTrips({}, 1, pagination.pageSize);
  }, []);

  function handleTableChange(nextPagination: TablePaginationConfig) {
    const nextPageNum = nextPagination.current ?? 1;
    const nextPageSize = nextPagination.pageSize ?? pagination.pageSize;
    void loadTrips(query, nextPageNum, nextPageSize);
  }

  return (
    <Card title="行程管理">
      <TripSearchForm
        onSearch={(values) => {
          setQuery(values);
          void loadTrips(values, 1, pagination.pageSize);
        }}
      />
      <Table
        columns={columns}
        dataSource={data}
        loading={loading}
        onChange={handleTableChange}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
        }}
        rowKey="id"
      />
    </Card>
  );
}
