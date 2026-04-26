import type { ReactNode } from "react";
import { Button, Popconfirm } from "antd";

type ConfirmSubmitButtonProps = {
  children: ReactNode;
  confirmTitle: string;
  disabled?: boolean;
  loading?: boolean;
  onConfirm: () => void | Promise<void>;
};

export default function ConfirmSubmitButton({
  children,
  confirmTitle,
  disabled,
  loading,
  onConfirm,
}: ConfirmSubmitButtonProps) {
  return (
    <Popconfirm okText="确认" cancelText="取消" title={confirmTitle} onConfirm={onConfirm}>
      <Button disabled={disabled} loading={loading} type="primary">
        {children}
      </Button>
    </Popconfirm>
  );
}
