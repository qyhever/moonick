import { useEffect, useState } from "react";
import { Card, Descriptions, Form, Input, InputNumber, Select, Space, Switch } from "antd";
import { useNavigate, useParams } from "react-router-dom";

import ConfirmSubmitButton from "../../components/ConfirmSubmitButton";
import {
  getAdminTripDetail,
  updateAdminTrip,
  type AdminTripDetail,
  type AdminTripUpdatePayload,
} from "./api";
import { getTripStatusText } from "../displayText";

const tripTypeOptions = [
  { label: "车找人", value: "driver_post" },
  { label: "人找车", value: "passenger_post" },
];

const statusOptions = [
  { label: "可约", value: "active" },
  { label: "已满", value: "full" },
  { label: "已关闭", value: "closed" },
  { label: "已过期", value: "expired", disabled: true },
];

function toFormValues(trip: AdminTripDetail): AdminTripUpdatePayload {
  return {
    tripType: trip.tripType,
    fromText: trip.fromText,
    toText: trip.toText,
    departureDate: trip.departureDate,
    departureTime: trip.departureTime,
    seatCount: trip.seatCount,
    priceAmount: trip.priceAmount,
    isPriceNegotiable: trip.isPriceNegotiable,
    contactWechat: trip.contactWechat,
    contactPhone: trip.contactPhone,
    remark: trip.remark,
    status: trip.status,
  };
}

export default function TripEditPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [trip, setTrip] = useState<AdminTripDetail | null>(null);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm<AdminTripUpdatePayload>();

  useEffect(() => {
    let active = true;

    async function load() {
      const data = await getAdminTripDetail(id);
      if (!active) {
        return;
      }
      setTrip(data);
      form.setFieldsValue(toFormValues(data));
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
      await updateAdminTrip(id, values);
      navigate(`/trips/${id}`);
    } finally {
      setSaving(false);
    }
  }

  if (!trip) {
    return <Card>加载中...</Card>;
  }

  const isExpired = trip.status === "expired";

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card title={`编辑行程 #${trip.id}`}>
        <Descriptions bordered column={2} style={{ marginBottom: 16 }}>
          <Descriptions.Item label="路线">{trip.fromText} → {trip.toText}</Descriptions.Item>
          <Descriptions.Item label="当前状态">{getTripStatusText(trip.status)}</Descriptions.Item>
          <Descriptions.Item label="出发时间">{trip.departureDate} {trip.departureTime}</Descriptions.Item>
          <Descriptions.Item label="联系方式">{trip.contactPhone || trip.contactWechat || "未填写"}</Descriptions.Item>
        </Descriptions>

        <Form form={form} layout="vertical">
          <Form.Item label="行程类型" name="tripType" rules={[{ required: true, message: "请选择行程类型" }]}>
            <Select disabled={isExpired} options={tripTypeOptions} />
          </Form.Item>

          <Form.Item label="出发地" name="fromText" rules={[{ required: true, message: "请输入出发地" }]}>
            <Input disabled={isExpired} />
          </Form.Item>

          <Form.Item label="目的地" name="toText" rules={[{ required: true, message: "请输入目的地" }]}>
            <Input disabled={isExpired} />
          </Form.Item>

          <Space size={16} style={{ display: "flex" }} wrap>
            <Form.Item
              label="出发日期"
              name="departureDate"
              rules={[{ required: true, message: "请输入出发日期" }]}
              style={{ flex: 1, minWidth: 200 }}
            >
              <Input disabled={isExpired} placeholder="YYYY-MM-DD" />
            </Form.Item>

            <Form.Item
              label="出发时间"
              name="departureTime"
              rules={[{ required: true, message: "请输入出发时间" }]}
              style={{ flex: 1, minWidth: 200 }}
            >
              <Input disabled={isExpired} placeholder="HH:mm" />
            </Form.Item>
          </Space>

          <Space size={16} style={{ display: "flex" }} wrap>
            <Form.Item
              label="座位数"
              name="seatCount"
              rules={[{ required: true, message: "请输入座位数" }]}
              style={{ flex: 1, minWidth: 200 }}
            >
              <InputNumber disabled={isExpired} min={1} precision={0} style={{ width: "100%" }} />
            </Form.Item>

            <Form.Item
              label="费用金额"
              name="priceAmount"
              rules={[{ required: true, message: "请输入费用金额" }]}
              style={{ flex: 1, minWidth: 200 }}
            >
              <InputNumber disabled={isExpired} min={0} precision={2} style={{ width: "100%" }} />
            </Form.Item>
          </Space>

          <Form.Item label="允许议价" name="isPriceNegotiable" valuePropName="checked">
            <Switch disabled={isExpired} />
          </Form.Item>

          <Form.Item label="微信号" name="contactWechat">
            <Input disabled={isExpired} />
          </Form.Item>

          <Form.Item label="手机号" name="contactPhone">
            <Input disabled={isExpired} />
          </Form.Item>

          <Form.Item label="备注" name="remark">
            <Input.TextArea disabled={isExpired} rows={4} />
          </Form.Item>

          <Form.Item label="行程状态" name="status" rules={[{ required: true, message: "请选择状态" }]}>
            <Select disabled={isExpired} options={statusOptions} />
          </Form.Item>

          <ConfirmSubmitButton
            confirmTitle="确认保存修改吗？"
            disabled={isExpired}
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
