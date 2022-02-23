package eventdefintions

stripe_customer_created: #Def & {
	description: "Sent when a customer is created"
	schema: {
		name: "stripe/customer.created"
		data: {
			livemode: bool
			// The unique event ID from stripe.
			id: string
			data: {
				object: {
					default_source?: string
					delinquent:      bool
					invoice_prefix:  string
					invoice_settings: {
						custom_fields?: [...{name: string, value: string}]
						default_payment_method?: string
						footer?:                 string
					}
					livemode: bool
					metadata: {}
					preferred_locales: [...string]
					id:        string
					name?:     string
					shipping:  _
					balance:   int
					currency?: string
					created:   int
					address?: {
						city:        string
						country:     string
						line1:       string
						line2?:      string
						postal_code: string
						state:       string
					}
					description: string
					discount?: {
						id:    string
						start: int
						end:   int
						...
					}
					email?:                string
					next_invoice_sequence: int
					phone?:                string
					tax_exempt:            string
					object:                string
				}
			}
			request: {
				id:              string
				idempotency_key: string
			}
			pending_webhooks: int
			type:             string
			object:           string
			api_version:      string
			created:          int
		}
		user?: {
			email?: string
		}
	}

	examples: [
		"""
			{
			  "id": "evt_1KWTPTJKEym2H9Vg1CGtbZc8",
			  "object": "event",
			  "api_version": "2020-08-27",
			  "created": 1645655651,
			  "data": {
			    "object": {
			      "id": "cus_LCtB7H0tTFBxEn",
			      "object": "customer",
			      "address": null,
			      "balance": 0,
			      "created": 1645655651,
			      "currency": null,
			      "default_source": null,
			      "delinquent": false,
			      "description": "(created by Stripe CLI)",
			      "discount": null,
			      "email": null,
			      "invoice_prefix": "DFF92B48",
			      "invoice_settings": {
				"custom_fields": null,
				"default_payment_method": null,
				"footer": null
			      },
			      "livemode": false,
			      "metadata": {},
			      "name": null,
			      "next_invoice_sequence": 1,
			      "phone": null,
			      "preferred_locales": [],
			      "shipping": null,
			      "tax_exempt": "none"
			    }
			  },
			  "livemode": false,
			  "pending_webhooks": 2,
			  "request": {
			    "id": "req_kgeJvoqQgzYOz7",
			    "idempotency_key": "a901cf97-c405-4536-8089-c2af77d57d10"
			  },
			  "type": "customer.created"
			}
			""",
	]
}

stripe_charge_succeeded: #Def & {
	description: "Sent when a charge completes successfully in your account"
	schema: {
		name: "stripe/charge.succeeded"
		data: {
			id:          string
			type:        "charge.succeeded"
			object:      string
			api_version: string
			created:     int
			data: {
				object: {
					amount_captured:             int
					receipt_number:              _
					receipt_url:                 string
					source_transfer:             _
					statement_descriptor_suffix: _
					transfer_data:               _
					amount:                      int
					dispute:                     _
					disputed:                    bool
					fraud_details: {
						stripe_report?: "fraudulent"
						user_report?:   "fraudulent" | "safe"
					}
					livemode: bool
					metadata: {}
					// The ID of the order for this charge, if one eixsts.
					order:    string | null
					shipping: _
					billing_details: {
						address: {
							city:        string | null
							country:     string | null
							line1:       string | null
							line2:       string | null
							postal_code: string | null
							state:       string | null
						}
						email: string | null
						name:  string | null
						phone: string | null
					}
					// The stripe ID of the customer for this charge, if one exists.
					customer:            string | null
					payment_method:      string
					transfer_group:      _
					amount_refunded:     int
					refunded:            bool
					review:              string | null
					created:             int
					balance_transaction: string | null
					on_behalf_of:        _
					outcome: {
						seller_message: string
						type:           string
						network_status: string
						reason:         _
						risk_level:     string
						risk_score:     int
					}
					statement_descriptor:            _
					status:                          string
					application:                     _
					calculated_statement_descriptor: string
					captured:                        bool
					// The error message explaining the reason for failure, if failed
					failure_message: string | null
					receipt_email:   _
					refunds: {
						total_count: int
						url:         string
						object:      string
						data: [...]
						has_more: bool
					}
					application_fee_amount: _
					object:                 string
					paid:                   bool
					payment_intent:         _
					id:                     string
					currency:               string
					description:            string
					destination:            _
					failure_code:           _
					invoice:                _
					payment_method_details: {
						card: {
							checks: {
								address_line1_check:       _
								address_postal_code_check: _
								cvc_check:                 _
							}
							country:        string
							exp_month:      int
							last4:          string
							network:        string
							three_d_secure: _
							brand:          string
							exp_year:       int
							fingerprint:    string
							funding:        string
							installments:   _
							wallet:         _
						}
						type: string
					}
					source: {
						address_city:  string | null
						country:       string
						dynamic_last4: string | null
						exp_month:     int
						funding:       string
						metadata: {}
						address_zip:         string | null
						customer:            string | null
						cvc_check:           string | null
						object:              string
						address_country:     string | null
						brand:               string
						exp_year:            int
						name:                string | null
						fingerprint:         string
						last4:               string
						id:                  string
						address_line1:       string | null
						address_line1_check: string | null
						address_line2:       string | null
						address_state:       string | null
						address_zip_check:   string | null
						tokenization_method: string | null
					}
					application_fee: _
				}
			}
			livemode:         bool
			pending_webhooks: int
			request: {
				id:              string
				idempotency_key: string
			}
		}
		user?: {
			email?: string
		}
	}

	examples: [
		"""
			{
			  "id": "evt_3KWSdjJKEym2H9Vg0VTPKkfD",
			  "object": "event",
			  "api_version": "2020-08-27",
			  "created": 1645652691,
			  "data": {
			    "object": {
			      "id": "ch_3KWSdjJKEym2H9Vg0qAid6Vb",
			      "object": "charge",
			      "amount": 100,
			      "amount_captured": 0,
			      "amount_refunded": 0,
			      "application": null,
			      "application_fee": null,
			      "application_fee_amount": null,
			      "balance_transaction": null,
			      "billing_details": {
				"address": {
				  "city": null,
				  "country": null,
				  "line1": null,
				  "line2": null,
				  "postal_code": null,
				  "state": null
				},
				"email": null,
				"name": null,
				"phone": null
			      },
			      "calculated_statement_descriptor": "INNGEST.COM",
			      "captured": false,
			      "created": 1645652691,
			      "currency": "usd",
			      "customer": null,
			      "description": "(created by Stripe CLI)",
			      "destination": null,
			      "dispute": null,
			      "disputed": false,
			      "failure_code": null,
			      "failure_message": null,
			      "fraud_details": {},
			      "invoice": null,
			      "livemode": false,
			      "metadata": {},
			      "on_behalf_of": null,
			      "order": null,
			      "outcome": {
				"network_status": "approved_by_network",
				"reason": null,
				"risk_level": "normal",
				"risk_score": 31,
				"seller_message": "Payment complete.",
				"type": "authorized"
			      },
			      "paid": true,
			      "payment_intent": null,
			      "payment_method": "card_1KWSdjJKEym2H9Vg5nVjO98Q",
			      "payment_method_details": {
				"card": {
				  "brand": "visa",
				  "checks": {
				    "address_line1_check": null,
				    "address_postal_code_check": null,
				    "cvc_check": null
				  },
				  "country": "US",
				  "exp_month": 2,
				  "exp_year": 2023,
				  "fingerprint": "Te4OI5BJL6MnK9Ey",
				  "funding": "credit",
				  "installments": null,
				  "last4": "4242",
				  "network": "visa",
				  "three_d_secure": null,
				  "wallet": null
				},
				"type": "card"
			      },
			      "receipt_email": null,
			      "receipt_number": null,
			      "receipt_url": "https://pay.stripe.com/receipts/acct_1IVczWJKEym2H9Vg/ch_3KWSdjJKEym2H9Vg0qAid6Vb/rcpt_LCsOjvrEm5K2yuZe50LNfm4AeiblR3e",
			      "refunded": false,
			      "refunds": {
				"object": "list",
				"data": [],
				"has_more": false,
				"total_count": 0,
				"url": "/v1/charges/ch_3KWSdjJKEym2H9Vg0qAid6Vb/refunds"
			      },
			      "review": null,
			      "shipping": null,
			      "source": {
				"id": "card_1KWSdjJKEym2H9Vg5nVjO98Q",
				"object": "card",
				"address_city": null,
				"address_country": null,
				"address_line1": null,
				"address_line1_check": null,
				"address_line2": null,
				"address_state": null,
				"address_zip": null,
				"address_zip_check": null,
				"brand": "Visa",
				"country": "US",
				"customer": null,
				"cvc_check": null,
				"dynamic_last4": null,
				"exp_month": 2,
				"exp_year": 2023,
				"fingerprint": "Te4OI5BJL6MnK9Ey",
				"funding": "credit",
				"last4": "4242",
				"metadata": {},
				"name": null,
				"tokenization_method": null
			      },
			      "source_transfer": null,
			      "statement_descriptor": null,
			      "statement_descriptor_suffix": null,
			      "status": "succeeded",
			      "transfer_data": null,
			      "transfer_group": null
			    }
			  },
			  "livemode": false,
			  "pending_webhooks": 1,
			  "request": {
			    "id": "req_fuQl3aU6rajYFa",
			    "idempotency_key": "717c61b7-0203-4d47-944a-142bc61cdb44"
			  },
			  "type": "charge.succeeded"
			}
			""",
	]
}
