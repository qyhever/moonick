import { useEffect, useState } from "react";
import { Card, Input, Space, Table } from "antd";
import { Link } from "react-router-dom";
import type { ColumnsType } from "antd/es/table";

import { getAdminUsers, type AdminUserSummary } from "./api";

const columns: ColumnsType<AdminUserSummary> = [
  { title: "ID", dataIndex: "id", width: 90 },
  { title: "昵称", dataIndex: "nickname" },
  { title: "手机号", dataIndex: "phone" },
  { title: "状态", dataIndex: "status" },
  {
    title: "操作",
    render: (_, record) => <Link to={`/users/${record.id}`}>详情</Link>,
  },
];

export default function UserListPage() {
  const [keyword, setKeyword] = useState("");
  const [data, setData] = useState<AdminUserSummary[]>([]);
  const [loading, setLoading] = useState(false);

  async function loadUsers(nextKeyword = "") {
    setLoading(true);
    try {
      const response = await getAdminUsers({ keyword: nextKeyword });
      setData(response.items);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadUsers();
  }, []);

  return (
    <Card title="用户管理">
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          allowClear
          enterButton="搜索"
          onSearch={(value) => {
            setKeyword(value);
            void loadUsers(value);
          }}
          placeholder="按昵称或手机号搜索"
          value={keyword}
          onChange={(event) => setKeyword(event.target.value)}
        />
      </Space>
      <Table columns={columns} dataSource={data} loading={loading} pagination={false} rowKey="id" />
    </Card>
  );
}
