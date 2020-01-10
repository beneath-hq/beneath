import React, { FC } from 'react';
import { Typography } from "@material-ui/core";

interface Props {
  billing_period: string
  description: string | null
}

const PaymentsByAnarchism: FC<Props> = ({ billing_period, description }) => {
  return (
    <div>
      <Typography variant="body1">Current billing plan: {description}</Typography>
      <Typography variant="body1">Current billing period: {billing_period}</Typography>
      <p> You are on a special plan... you don't have to pay! </p>
    </div>
  )
}

export default PaymentsByAnarchism