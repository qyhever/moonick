import { useEffect, useState } from "react";
import { Card, Input, Space, Table } from "antd";
import { Link } from "react-router-dom";
import type { ColumnsType } from "antd/es/table";
import type { TablePaginationConfig } from "antd/es/table";

import { getAdminUsers, type AdminUserSummary } from "./api";
import { getUserStatusText } from "../displayText";
import { formatDateTime } from "../../lib/dateTime";

const columns: ColumnsType<AdminUserSummary> = [
  { title: "ID", dataIndex: "id", width: 90 },
  { title: "昵称", dataIndex: "nickname" },
  { title: "手机号", dataIndex: "phone" },
  {
    title: "注册时间",
    dataIndex: "createdAt",
    render: (value: string) => formatDateTime(value),
  },
  {
    title: "状态",
    dataIndex: "status",
    render: (value: string) => getUserStatusText(value),
  },
  {
    title: "操作",
    render: (_, record) => <Link to={`/users/${record.id}`}>详情</Link>,
  },
];

export default function UserListPage() {
  const [keyword, setKeyword] = useState("");
  const [data, setData] = useState<AdminUserSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  async function loadUsers(nextPageNum = 1, nextPageSize = pagination.pageSize, nextKeyword = keyword) {
    setLoading(true);
    try {
      const response = await getAdminUsers({
        keyword: nextKeyword,
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
    void loadUsers(1, pagination.pageSize, "");
  }, []);

  function handleTableChange(nextPagination: TablePaginationConfig) {
    const nextPageNum = nextPagination.current ?? 1;
    const nextPageSize = nextPagination.pageSize ?? pagination.pageSize;
    void loadUsers(nextPageNum, nextPageSize, keyword);
  }

  return (
    <Card title="用户管理">
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          allowClear
          enterButton="搜索"
          onSearch={(value) => {
            setKeyword(value);
            void loadUsers(1, pagination.pageSize, value);
          }}
          placeholder="按昵称或手机号搜索"
          value={keyword}
          onChange={(event) => setKeyword(event.target.value)}
        />
      </Space>
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
