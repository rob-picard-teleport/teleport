// Teleport
// Copyright (C) 2025 Gravitational, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package vnet

import (
	"context"
	"crypto"
	"crypto/rsa"
	"io"

	"github.com/gravitational/trace"

	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
)

type remoteSigner struct {
	pub         crypto.PublicKey
	sendRequest func(context.Context, *vnetv1.SignRequest) ([]byte, error)
}

// Public implements [crypto.Signer.Public] and returns the public key
// associated with the signer.
func (s *remoteSigner) Public() crypto.PublicKey {
	return s.pub
}

// Sign implements [crypto.Signer.Sign] and issues a signature over digest for
// the associated app.
func (s *remoteSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	req := &vnetv1.SignRequest{
		Digest: digest,
	}
	switch opts.HashFunc() {
	case 0:
		req.Hash = vnetv1.Hash_HASH_NONE
	case crypto.SHA256:
		req.Hash = vnetv1.Hash_HASH_SHA256
	default:
		return nil, trace.BadParameter("unsupported signature hash func %v", opts.HashFunc())
	}
	if pssOpts, ok := opts.(*rsa.PSSOptions); ok {
		saltLen := int32(pssOpts.SaltLength)
		req.PssSaltLength = &saltLen
	}
	signature, err := s.sendRequest(context.TODO(), req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return signature, nil
}
