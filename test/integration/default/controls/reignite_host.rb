# frozen_string_literal: true

control "reignite_host" do
    desc "Verifies that the machine is a reignite host"
  
    describe file("/tmp/metadata") do
      its('size') { should be >= 10 }
    end

    describe interface('bond0') do
      it { should be_up }
    end

    describe kernel_module('kvm') do
      it { should be_loaded }
    end
  end
  
  