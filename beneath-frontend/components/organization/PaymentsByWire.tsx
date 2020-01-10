import React, { FC } from 'react';
import { makeStyles, TextField, Typography, Button } from "@material-ui/core";

interface Props {
  billing_period: string
  description: string | null
}

const PaymentsByWire: FC<Props> = ({ billing_period, description }) => {
  return (
    <div>
      <Typography variant="body1">Current billing plan: {description}</Typography>
      <Typography variant="body1">Current billing period: {billing_period}</Typography>
      <p> You're paying by wire </p>
    </div>
  )
}

export default PaymentsByWire