import { Button, Form, Input, Select, Space } from "antd";

type TripSearchValues = {
  keyword?: string;
  status?: string;
};

type TripSearchFormProps = {
  onSearch: (values: TripSearchValues) => void;
};

export default function TripSearchForm({ onSearch }: TripSearchFormProps) {
  const [form] = Form.useForm<TripSearchValues>();

  return (
    <Form
      form={form}
      layout="inline"
      onFinish={onSearch}
      style={{ marginBottom: 16, rowGap: 12 }}
    >
      <Form.Item name="keyword">
        <Input placeholder="按路线关键字搜索" />
      </Form.Item>
      <Form.Item name="status">
        <Select
          allowClear
          options={[
            { label: "可约", value: "active" },
            { label: "已满", value: "full" },
            { label: "已关闭", value: "closed" },
            { label: "已过期", value: "expired" },
          ]}
          placeholder="状态"
          style={{ width: 140 }}
        />
      </Form.Item>
      <Space>
        <Button htmlType="submit" type="primary">
          查询
        </Button>
        <Button
          onClick={() => {
            form.resetFields();
            onSearch({});
          }}
        >
          重置
        </Button>
      </Space>
    </Form>
  );
}
