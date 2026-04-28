import { useEffect, useState } from "react";
import { Card, Descriptions, Space, Tag } from "antd";
import { Link, useParams } from "react-router-dom";

import { getAdminTripDetail, type AdminTripDetail } from "./api";

function formatPrice(trip: AdminTripDetail) {
  if (trip.isPriceNegotiable && trip.priceAmount > 0) {
    return `${trip.priceAmount} 元（可议价）`;
  }

  if (trip.isPriceNegotiable) {
    return "面议";
  }

  return `${trip.priceAmount} 元`;
}

export default function TripDetailPage() {
  const { id = "" } = useParams();
  const [trip, setTrip] = useState<AdminTripDetail | null>(null);

  useEffect(() => {
    let active = true;

    async function load() {
      const data = await getAdminTripDetail(id);
      if (active) {
        setTrip(data);
      }
    }

    void load();
    return () => {
      active = false;
    };
  }, [id]);

  if (!trip) {
    return <Card>加载中...</Card>;
  }

  return (
    <Space direction="vertical" size={16} style={{ width: "100%" }}>
      <Card
        extra={trip.status === "expired" ? null : <Link to={`/trips/${trip.id}/edit`}>编辑行程信息</Link>}
        title={`行程详情 #${trip.id}`}
      >
        <Descriptions bordered column={2}>
          <Descriptions.Item label="类型">{trip.tripType}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag>{trip.status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="起点">{trip.fromText}</Descriptions.Item>
          <Descriptions.Item label="终点">{trip.toText}</Descriptions.Item>
          <Descriptions.Item label="出发日期">{trip.departureDate}</Descriptions.Item>
          <Descriptions.Item label="出发时间">{trip.departureTime}</Descriptions.Item>
          <Descriptions.Item label="人数">{trip.seatCount}</Descriptions.Item>
          <Descriptions.Item label="费用">{formatPrice(trip)}</Descriptions.Item>
          <Descriptions.Item label="微信号">{trip.contactWechat || "未填写"}</Descriptions.Item>
          <Descriptions.Item label="手机号">{trip.contactPhone || "未填写"}</Descriptions.Item>
          <Descriptions.Item label="备注" span={2}>{trip.remark || "未填写"}</Descriptions.Item>
          <Descriptions.Item label="发布时间">{trip.createdAt}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{trip.updatedAt}</Descriptions.Item>
        </Descriptions>
      </Card>
    </Space>
  );
}
