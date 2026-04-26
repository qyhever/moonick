import { useEffect, useState } from "react";
import { Card, Descriptions, Form, Select, Space } from "antd";
import { useNavigate, useParams } from "react-router-dom";

import ConfirmSubmitButton from "../../components/ConfirmSubmitButton";
import { getAdminTripDetail, updateAdminTrip, type AdminTripDetail } from "./api";

type FormValues = {
  status: "active" | "full" | "closed";
};

export default function TripEditPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [trip, setTrip] = useState<AdminTripDetail | null>(null);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm<FormValues>();

  useEffect(() => {
    let active = true;

    async function load() {
      const data = await getAdminTripDetail(id);
      if (!active) {
        return;
      }
      setTrip(data);
      if (data.status !== "expired") {
        form.setFieldsValue({
          status: data.status,
        });
      }
    }

    void load();
    return () => {
      active = false;
    };
  }, [form, id]);

  async function handleConfirm() {
    const values = await form.validateFields();
    setSaving(true);
    try {
      await updateAdminTrip(id, values.status);
      navigate(`/trips/${id}`);
    } finally {
      setSaving(false);
    }
  }

  if (!trip) {
    return <Card>加载中...</Card>;
  }

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title={`编辑行程 #${trip.id}`}>
        <Descriptions bordered column={2} style={{ marginBottom: 16 }}>
          <Descriptions.Item label="路线">{trip.fromText} → {trip.toText}</Descriptions.Item>
          <Descriptions.Item label="当前状态">{trip.status}</Descriptions.Item>
          <Descriptions.Item label="出发时间">{trip.departureDate} {trip.departureTime}</Descriptions.Item>
          <Descriptions.Item label="联系方式">{trip.contactPhone || trip.contactWechat || "未填写"}</Descriptions.Item>
        </Descriptions>

        <Form form={form} layout="vertical">
          <Form.Item label="行程状态" name="status" rules={[{ required: true, message: "请选择状态" }]}>
            <Select
              disabled={trip.status === "expired"}
              options={[
                { label: "可约", value: "active" },
                { label: "已满", value: "full" },
                { label: "已关闭", value: "closed" },
              ]}
            />
          </Form.Item>

          <ConfirmSubmitButton
            confirmTitle="确认保存修改吗？"
            disabled={trip.status === "expired"}
            loading={saving}
            onConfirm={() => void handleConfirm()}
          >
            保存
          </ConfirmSubmitButton>
        </Form>
      </Card>
    </Space>
  );
}
